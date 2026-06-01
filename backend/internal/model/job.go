package model

import (
	"time"

	"gorm.io/gorm"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	ID           string         `gorm:"primaryKey" json:"id"`
	Status       JobStatus      `gorm:"index" json:"status"`
	Progress     int            `json:"progress"`
	Total        int            `json:"total"`
	CurrentHS    string         `json:"current_hs"`
	Message      string         `json:"message"`
	ErrorMsg     string         `json:"error_msg,omitempty"`
	SourcePath   string         `json:"-"`
	TemplatePath string         `json:"-"`
	OutputDir    string         `json:"-"`
	OutputBatch  string         `json:"output_batch,omitempty"`
	ZipPath      string         `json:"-"`
	ZipFileName  string         `json:"zip_file_name,omitempty"`
	FileCount    int            `json:"file_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type OutputFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	JobID     string    `gorm:"index" json:"job_id"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"-"`
	HSCode    string    `json:"hs_code"`
	CINo      string    `json:"ci_no"`
	RowCount  int       `json:"row_count"`
	CreatedAt time.Time `json:"created_at"`
}

type DataRow struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	JobID       string `gorm:"index" json:"job_id"`
	RowNum      int    `json:"row_num"`
	No          string `json:"no"`
	RecordID    string `json:"record_id"`
	CINo        string `json:"ci_no"`
	ContactNo   string `json:"contact_no"`
	ContainerNo string `json:"container_no"`
	SealNo      string `json:"seal_no"`
	Contract    string `json:"contract"`
	Item        string `json:"item"`
	ModNo       string `json:"mod_no"`
	CaseNo      string `json:"case_no"`
	PartNo      string `json:"part_no"`
	HSCode      string `json:"hs_code"`
	DescEN      string `json:"desc_en"`
	DescRU      string `json:"desc_ru"`
	Qty         string `json:"qty"`
	UnitPrice   string `json:"unit_price"`
	Freight1    string `json:"freight1"`
	Insurance   string `json:"insurance"`
	UnitCIP     string `json:"unit_cip"`
	NW          string `json:"nw"`
	TotalNW     string `json:"total_nw"`
	GW          string `json:"gw"`
	TotalGW     string `json:"total_gw"`
	TotalCIP    string `json:"total_cip"`
	Length      string `json:"length"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	Volume      string `json:"volume"`
	Port        string `json:"port"`
	Type        string `json:"type"`
	Pkgs        string `json:"pkgs"`
	TotalNWKgs  string `json:"total_nw_kgs"`
	TotalGWKgs  string `json:"total_gw_kgs"`
	PkgDim      string `json:"pkg_dim"`
	PkgMark     string `json:"pkg_mark"`
	Manufacturer string `json:"manufacturer"`
	TradeMark   string `json:"trade_mark"`
}
