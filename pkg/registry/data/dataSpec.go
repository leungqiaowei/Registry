package data

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type FileType string

const (
	FileTypeFile   FileType = "file"
	FileTypeFolder FileType = "folder"
)

// DataSpec 文件元数据结构
type DataSpec struct {
	FileName    string    `json:"file_name"`
	Tag         string    `json:"tag"`
	Type        FileType  `json:"type"`
	IsPermanent bool      `json:"is_permanent"`
	FilePath    string    `json:"file_path"`
	StorageTime time.Time `json:"storage_time"`
	Owner       string    `json:"owner"`
	Size        int64     `json:"size"`
	FileHash    string    `json:"file_hash"`
}

type DataSpecList struct {
	DataSpecIndex map[string]DataSpec
	mutex         sync.RWMutex
}

func NewDataSpecList() *DataSpecList {
	return &DataSpecList{DataSpecIndex: make(map[string]DataSpec)}
}

func NewDataSpec(fileName, tag, owner string, fileType FileType, isPermanent bool) *DataSpec {
	return &DataSpec{
		FileName:    fileName,
		Tag:         tag,
		Type:        fileType,
		FilePath:    "",
		IsPermanent: isPermanent,
		StorageTime: time.Now(),
		FileHash:    "",
		Owner:       owner,
		Size:        0,
	}
}

func generateFilePath(dest, fileName, tag string) string {
	return filepath.Join(dest, fmt.Sprintf("%s_%s", fileName, tag))
}

func generateFolderPath(dest, fileName string) string {
	return filepath.Join(dest, fmt.Sprintf("%s", fileName))
}

func getKey(fileName, tag string) string {
	return fmt.Sprintf("%s_%s", fileName, tag)
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// SaveFile 保存文件到指定目录，并更新 DataSpecIndex
func (dsl *DataSpecList) SaveFile(dest, fileName, tag, owner string, isPermanent bool, data io.Reader) (*DataSpec, error) {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()

	filePath := generateFilePath(dest, fileName, tag)
	if fileExists(filePath) {
		return nil, fmt.Errorf("file already exists: %s_%s", fileName, tag)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	written, err := io.Copy(file, data)
	if err != nil {
		return nil, err
	}

	fileHash, err := calculateFileHash(filePath)
	if err != nil {
		return nil, err
	}

	d := DataSpec{
		FileName:    fileName,
		Tag:         tag,
		Type:        FileTypeFile,
		FilePath:    filePath,
		IsPermanent: isPermanent,
		Owner:       owner,
		Size:        written,
		StorageTime: time.Now(),
		FileHash:    fileHash,
	}
	key := getKey(fileName, tag)
	dsl.DataSpecIndex[key] = d

	return &d, nil
}

func (dsl *DataSpecList) SaveFolder(dest, fileName, tag, owner string, isPermanent bool) (*DataSpec, error) {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()

	filePath := generateFolderPath(dest, fileName)
	d := DataSpec{
		FileName:    fileName,
		Tag:         tag,
		Type:        FileTypeFolder,
		FilePath:    filePath,
		IsPermanent: isPermanent,
		Owner:       owner,
		Size:        0,
		StorageTime: time.Now(),
		FileHash:    "",
	}
	key := getKey(fileName, tag)
	dsl.DataSpecIndex[key] = d
	return &d, nil
}

func (dsl *DataSpecList) GetFilePath(fileName, tag string) string {
	dsl.mutex.RLock()
	defer dsl.mutex.RUnlock()
	key := getKey(fileName, tag)
	dataSpec, exists := dsl.DataSpecIndex[key]
	if !exists {
		return ""
	}
	return dataSpec.FilePath
}

func (dsl *DataSpecList) GetFolderPath(dest, fileName, tag string) string {
	dsl.mutex.RLock()
	defer dsl.mutex.RUnlock()
	key := getKey(fileName, tag)
	dataSpec, exists := dsl.DataSpecIndex[key]
	if !exists {
		return ""
	}
	folderPath, err := filepath.Rel(dest, dataSpec.FilePath)
	if err != nil {
		return dataSpec.FilePath
	}
	return folderPath
}

// LoadFile 从指定目录加载文件，并更新 DataSpecIndex
func (dsl *DataSpecList) LoadFile(fileName, tag string) (*os.File, error) {
	dsl.mutex.RLock()
	dataSpec, exists := dsl.DataSpecIndex[getKey(fileName, tag)]
	dsl.mutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("file not found: %s", getKey(fileName, tag))
	}
	return os.Open(dataSpec.FilePath)
}

// GetDataSpec 根据文件路径获取 DataSpec
func (dsl *DataSpecList) GetDataSpec(fileName, tag string) (*DataSpec, error) {
	dsl.mutex.RLock()
	defer dsl.mutex.RUnlock()
	dataSpec, exists := dsl.DataSpecIndex[getKey(fileName, tag)]
	if !exists {
		return nil, fmt.Errorf("file not found: %s_%s", fileName, tag)
	}
	copy := dataSpec
	return &copy, nil
}

// GetList 列出所有 DataSpec
func (dsl *DataSpecList) GetList() []DataSpec {
	dsl.mutex.RLock()
	defer dsl.mutex.RUnlock()
	list := make([]DataSpec, 0, len(dsl.DataSpecIndex))
	for _, dataSpec := range dsl.DataSpecIndex {
		list = append(list, dataSpec)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].FileName == list[j].FileName {
			return list[i].Tag < list[j].Tag
		}
		return list[i].FileName < list[j].FileName
	})
	return list
}

// DeleteFile 删除文件并更新 DataSpecIndex
func (dsl *DataSpecList) DeleteFile(fileName, tag string) error {
	dsl.mutex.Lock()
	defer dsl.mutex.Unlock()

	key := getKey(fileName, tag)
	d, exists := dsl.DataSpecIndex[key]
	if !exists {
		return fmt.Errorf("file not found: %s", key)
	}
	if err := os.RemoveAll(d.FilePath); err != nil {
		return err
	}
	delete(dsl.DataSpecIndex, key)
	return nil
}

// UpdateFile 更新文件并更新 DataSpecIndex
func (dsl *DataSpecList) UpdateFile(dest, fileName, tag, owner string, isPermanent bool, data io.Reader) (*DataSpec, error) {
	key := getKey(fileName, tag)
	if err := dsl.DeleteFile(fileName, tag); err != nil {
		return nil, fmt.Errorf("failed to delete old file %s: %w", key, err)
	}
	return dsl.SaveFile(dest, fileName, tag, owner, isPermanent, data)
}
