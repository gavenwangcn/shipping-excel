package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"shipping-excel/backend/internal/logx"
	"shipping-excel/backend/internal/service"
)

type Handler struct {
	jobs *service.JobService
}

func New(jobs *service.JobService) *Handler {
	return &Handler{jobs: jobs}
}

func (h *Handler) Register(r *gin.Engine) {
	r.Use(h.requestLogger())
	api := r.Group("/api")
	{
		api.POST("/upload", h.Upload)
		api.GET("/jobs", h.ListJobs)
		api.GET("/jobs/:id", h.GetJob)
		api.GET("/jobs/:id/files", h.ListFiles)
		api.GET("/jobs/:id/files/:fileId/download", h.DownloadFile)
		api.GET("/jobs/:id/download", h.DownloadZip)
		api.GET("/jobs/:id/data", h.ListData)
	}
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func (h *Handler) requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logx.HTTP(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), c.Errors.String())
	}
}

func (h *Handler) Upload(c *gin.Context) {
	source, err := c.FormFile("source")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传源数据 Excel 文件 (source)"})
		return
	}
	template, err := c.FormFile("template")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传模板 Excel 文件 (template)"})
		return
	}

	srcFile, err := source.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取源文件失败"})
		return
	}
	defer srcFile.Close()

	tplFile, err := template.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取模板文件失败"})
		return
	}
	defer tplFile.Close()

	job, err := h.jobs.CreateJob(srcFile, tplFile, source.Filename, template.Filename)
	if err != nil {
		logx.Warnf("upload rejected: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logx.Jobf(job.ID, "upload accepted source=%q template=%q", source.Filename, template.Filename)

	c.JSON(http.StatusOK, gin.H{
		"id":      job.ID,
		"status":  job.Status,
		"message": job.Message,
	})
}

func (h *Handler) GetJob(c *gin.Context) {
	job, err := h.jobs.GetJob(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, job)
}

func (h *Handler) ListJobs(c *gin.Context) {
	jobs, err := h.jobs.ListJobs(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) ListFiles(c *gin.Context) {
	files, err := h.jobs.ListOutputFiles(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}

func (h *Handler) DownloadFile(c *gin.Context) {
	jobID := c.Param("id")
	fileID, err := strconv.ParseUint(c.Param("fileId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件 ID"})
		return
	}

	path, name, err := h.jobs.GetOutputFilePath(jobID, uint(fileID))
	if err != nil {
		logx.Warnf("download file job=%s file_id=%d failed: %v", jobID, fileID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	logx.Infof("download file job=%s file_id=%d name=%q path=%q", jobID, fileID, name, path)

	c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	c.File(path)
}

func (h *Handler) DownloadZip(c *gin.Context) {
	path, name, err := h.jobs.GetZipPath(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Disposition", "attachment; filename=\""+name+"\"")
	c.Header("Content-Type", "application/zip")
	c.File(path)
}

func (h *Handler) ListData(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	rows, total, err := h.jobs.ListDataRows(c.Param("id"), page, pageSize)
	if err != nil {
		logx.Errorf("list_data job=%s failed: %v", c.Param("id"), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      rows,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
