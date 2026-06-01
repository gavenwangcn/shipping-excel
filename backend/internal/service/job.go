package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"shipping-excel/backend/internal/archive"
	"shipping-excel/backend/internal/excel"
	"shipping-excel/backend/internal/model"
)

type JobService struct {
	db      *gorm.DB
	dataDir string
}

func NewJobService(db *gorm.DB, dataDir string) *JobService {
	return &JobService{db: db, dataDir: dataDir}
}

func (s *JobService) CreateJob(sourceFile, templateFile io.Reader, sourceName, templateName string) (*model.Job, error) {
	jobID := uuid.New().String()
	jobDir := filepath.Join(s.dataDir, "jobs", jobID)
	uploadDir := filepath.Join(jobDir, "uploads")
	outputDir := filepath.Join(jobDir, "output")

	for _, dir := range []string{uploadDir, outputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	sourcePath := filepath.Join(uploadDir, sanitizeName(sourceName))
	templatePath := filepath.Join(uploadDir, sanitizeName(templateName))

	if err := saveFile(sourcePath, sourceFile); err != nil {
		return nil, fmt.Errorf("保存源文件失败: %w", err)
	}
	if err := saveFile(templatePath, templateFile); err != nil {
		return nil, fmt.Errorf("保存模板文件失败: %w", err)
	}

	if err := excel.ValidateSourceFile(sourcePath); err != nil {
		os.RemoveAll(jobDir)
		return nil, err
	}
	if err := excel.ValidateTemplateFile(templatePath); err != nil {
		os.RemoveAll(jobDir)
		return nil, err
	}

	job := &model.Job{
		ID:           jobID,
		Status:       model.JobStatusPending,
		Progress:     0,
		SourcePath:   sourcePath,
		TemplatePath: templatePath,
		OutputDir:    outputDir,
		Message:      "任务已创建，等待处理",
	}
	if err := s.db.Create(job).Error; err != nil {
		return nil, err
	}

	go s.processJob(jobID)
	return job, nil
}

func (s *JobService) processJob(jobID string) {
	var job model.Job
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		return
	}

	s.updateJob(&job, model.JobStatusProcessing, 0, 0, "", "正在解析源数据...")

	records, err := excel.ParseDataSheet(job.SourcePath)
	if err != nil {
		s.failJob(&job, err.Error())
		return
	}

	if len(records) == 0 {
		s.failJob(&job, "数据表中未找到有效数据行")
		return
	}

	s.persistDataRows(jobID, records)

	hsCodes := excel.UniqueHSCodesSorted(records)
	total := len(hsCodes)

	s.updateJob(&job, model.JobStatusProcessing, 0, total, "", fmt.Sprintf("共 %d 个 HS CODE 待生成", total))

	batchName := time.Now().Format("20060102_150405")
	batchDir := filepath.Join(job.OutputDir, batchName)
	if err := os.MkdirAll(batchDir, 0755); err != nil {
		s.failJob(&job, "创建输出目录失败: "+err.Error())
		return
	}
	job.OutputDir = batchDir
	job.OutputBatch = batchName
	s.db.Save(&job)

	lastHS := excel.GetLastHSCodeFromOutputDir(batchDir)
	currentHS := excel.NextHSCode(hsCodes, lastHS)
	done := 0

	templateData, err := os.ReadFile(job.TemplatePath)
	if err != nil {
		s.failJob(&job, "读取模板文件失败: "+err.Error())
		return
	}

	if lastHS != "" && currentHS == "" {
		done = total
	}

	for currentHS != "" {
		s.updateJob(&job, model.JobStatusProcessing, done, total, currentHS,
			fmt.Sprintf("正在生成 HS CODE: %s (%d/%d)", currentHS, done+1, total))

		filtered := excel.FilterByHSCode(records, currentHS)
		result, err := excel.GenerateFromTemplateBytes(templateData, batchDir, currentHS, filtered)
		if err != nil {
			s.failJob(&job, fmt.Sprintf("生成 HS CODE %s 失败: %s", currentHS, err.Error()))
			return
		}

		outFile := model.OutputFile{
			JobID:    jobID,
			FileName: result.FileName,
			FilePath: result.FilePath,
			HSCode:   result.HSCode,
			CINo:     result.CINo,
			RowCount: result.RowCount,
		}
		s.db.Create(&outFile)

		done++
		s.updateJob(&job, model.JobStatusProcessing, done, total, currentHS,
			fmt.Sprintf("已完成 HS CODE: %s", currentHS))

		lastHS = currentHS
		currentHS = excel.NextHSCode(hsCodes, lastHS)
	}

	if done == 0 {
		s.failJob(&job, "未生成任何报关 Excel 文件")
		return
	}

	s.updateJob(&job, model.JobStatusProcessing, done, total, "", "正在打包 ZIP 压缩包...")

	zipFileName := fmt.Sprintf("customs_%s.zip", batchName)
	zipPath := filepath.Join(filepath.Dir(batchDir), zipFileName)
	if err := archive.ZipDir(batchDir, zipPath); err != nil {
		s.failJob(&job, "打包 ZIP 失败: "+err.Error())
		return
	}

	job.ZipPath = zipPath
	job.ZipFileName = zipFileName
	job.FileCount = done
	s.db.Save(&job)

	s.updateJob(&job, model.JobStatusCompleted, total, total, "",
		fmt.Sprintf("全部完成，共生成 %d 个报关 Excel，已打包为 %s", done, zipFileName))
}

func (s *JobService) persistDataRows(jobID string, records []excel.DataRecord) {
	var rows []model.DataRow
	for _, r := range records {
		rows = append(rows, model.DataRow{
			JobID:    jobID,
			RowNum:   r.RowNum,
			CINo:     r.CINo,
			PartNo:   r.PartNo,
			HSCode:   r.HSCode,
			DescEN:   r.DescEN,
			DescRU:   r.DescRU,
			Qty:      fmt.Sprintf("%g", r.Qty),
			UnitPrice: fmt.Sprintf("%g", r.UnitPrice),
			Freight1: fmt.Sprintf("%g", r.Freight1),
			Insurance: fmt.Sprintf("%g", r.Insurance),
			Type:     r.Type,
		})
	}
	if len(rows) > 0 {
		s.db.CreateInBatches(rows, 100)
	}
}

func (s *JobService) GetJob(jobID string) (*model.Job, error) {
	var job model.Job
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

func (s *JobService) ListOutputFiles(jobID string) ([]model.OutputFile, error) {
	var files []model.OutputFile
	err := s.db.Where("job_id = ?", jobID).Order("hs_code asc").Find(&files).Error
	return files, err
}

func (s *JobService) GetZipPath(jobID string) (string, string, error) {
	var job model.Job
	if err := s.db.First(&job, "id = ?", jobID).Error; err != nil {
		return "", "", err
	}
	if job.ZipPath == "" {
		return "", "", fmt.Errorf("压缩包尚未生成")
	}
	if _, err := os.Stat(job.ZipPath); err != nil {
		return "", "", fmt.Errorf("压缩包不存在")
	}
	name := job.ZipFileName
	if name == "" {
		name = filepath.Base(job.ZipPath)
	}
	return job.ZipPath, name, nil
}

func (s *JobService) GetOutputFilePath(jobID string, fileID uint) (string, string, error) {
	var f model.OutputFile
	if err := s.db.Where("id = ? AND job_id = ?", fileID, jobID).First(&f).Error; err != nil {
		return "", "", err
	}
	if _, err := os.Stat(f.FilePath); err != nil {
		return "", "", fmt.Errorf("文件不存在")
	}
	return f.FilePath, f.FileName, nil
}

func (s *JobService) ListDataRows(jobID string, page, pageSize int) ([]model.DataRow, int64, error) {
	var total int64
	var rows []model.DataRow
	q := s.db.Model(&model.DataRow{}).Where("job_id = ?", jobID)
	q.Count(&total)
	offset := (page - 1) * pageSize
	err := q.Order("hs_code asc, row_num asc").Offset(offset).Limit(pageSize).Find(&rows).Error
	return rows, total, err
}

func (s *JobService) ListJobs(limit int) ([]model.Job, error) {
	var jobs []model.Job
	err := s.db.Order("created_at desc").Limit(limit).Find(&jobs).Error
	return jobs, err
}

func (s *JobService) updateJob(job *model.Job, status model.JobStatus, progress, total int, currentHS, message string) {
	job.Status = status
	job.Progress = progress
	job.Total = total
	job.CurrentHS = currentHS
	job.Message = message
	job.UpdatedAt = time.Now()
	s.db.Save(job)
}

func (s *JobService) failJob(job *model.Job, errMsg string) {
	job.Status = model.JobStatusFailed
	job.ErrorMsg = errMsg
	job.Message = "任务失败"
	job.UpdatedAt = time.Now()
	s.db.Save(job)
}

func saveFile(path string, reader io.Reader) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func sanitizeName(name string) string {
	name = filepath.Base(name)
	replacer := strings.NewReplacer("..", "_")
	return replacer.Replace(name)
}
