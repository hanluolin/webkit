package main

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io"
	fs2 "io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/hanluolin/webkit/statik"
	"github.com/rakyll/statik/fs"
)

func main() {
	fmt.Println("webkit v1.1.0")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("请输入项目名称（默认'test_webkit'）：")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	projectName := strings.ReplaceAll(input, "\n", "")
	projectName = strings.ReplaceAll(input, "\r", "")
	if projectName == "" {
		projectName = "test_webkit"
	}

	fmt.Print("请输入项目路径（默认'./'）：")
	input, err = reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	objPath := strings.ReplaceAll(input, "\n", "")
	objPath = strings.ReplaceAll(input, "\r", "")
	objPath = filepath.Join(objPath, projectName)

	var dbType string
	for {
		fmt.Print("请选择数据库（pg、dm，默认pg）：")
		input, err = reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		dbType = strings.ReplaceAll(input, "\n", "")
		dbType = strings.ReplaceAll(input, "\r", "")
		if dbType == "" {
			dbType = "pg"
		}
		if strings.Contains(dbType, "pg") {
			dbType = "pg"
		}
		if strings.Contains(dbType, "dm") {
			dbType = "dm"
		}
		if dbType == "pg" || dbType == "dm" {
			break
		} else {
			fmt.Println("数据库错误，请输入pg、dm中的一项")
		}
	}

	fileSystem, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}
	if err = fs.Walk(fileSystem, "/", func(path string, info fs2.FileInfo, err error) error {
		if info.IsDir() {
			if err = os.MkdirAll(filepath.Join(objPath, path), 0755); err != nil {
				return fmt.Errorf("文件夹创建失败: %+v", err)
			}
		} else {
			destinationFile := filepath.Join(objPath, path)
			out, err := os.Create(destinationFile)
			if err != nil {
				return fmt.Errorf("failed to create destination file %s: %w", destinationFile, err)
			}
			defer out.Close()

			file, err := fs.ReadFile(fileSystem, path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			if info.Name() != "go.sum" {
				if info.Name() == "go.mod" {
					file = bytes.ReplaceAll(file, []byte("module webkit"), []byte("module "+projectName))
				} else {
					file = bytes.ReplaceAll(file, []byte("webkit/"), []byte(projectName+"/"))
				}
			}
			if info.Name() == "config.go" {
				switch dbType {
				case "dm":
					file = bytes.ReplaceAll(file, []byte(`Type: GetEnvString("DB_TYPE", "pg"),`), []byte(`Type: GetEnvString("DB_TYPE", "dm"),`))
					file = bytes.ReplaceAll(file, []byte(`Conn: GetEnvString("DB_CONN", "host=127.0.0.1 port=5432 user=cella dbname=test password=111111"),`), []byte(`Conn: GetEnvString("DB_CONN", "dm://username:password@ip:port?schema=dbname"),`))
				}
			}

			if _, err = io.Copy(out, bytes.NewReader(file)); err != nil {
				return fmt.Errorf("failed to copy content to destination file %s: %w", destinationFile, err)
			}
		}
		return nil
	}); err != nil {
		log.Println("error", err)
	}

	log.Println("Successfully generated.")
}

// copyEmbedDir 复制嵌入的目录及其子目录和文件到目标位置
func copyEmbedDir(embedFS embed.FS, embedDir, destinationDir string) error {
	// 读取嵌入的目录
	dirEntries, err := embedFS.ReadDir(embedDir)
	if err != nil {
		return fmt.Errorf("failed to read embedded directory: %w", err)
	}

	// 创建目标目录
	err = os.MkdirAll(destinationDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 遍历目录中的文件和子目录
	for _, entry := range dirEntries {
		// 构建完整路径
		entryPath := filepath.Join(embedDir, entry.Name())
		destinationPath := filepath.Join(destinationDir, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err = copyEmbedDir(embedFS, entryPath, destinationPath); err != nil {
				return fmt.Errorf("failed to copy embedded subdirectory %s: %w", entryPath, err)
			}
		} else {
			// 复制文件
			if err = copyEmbedFile(embedFS, entryPath, destinationPath); err != nil {
				return fmt.Errorf("failed to copy embedded file %s: %w", entryPath, err)
			}
		}
	}

	return nil
}

// copyEmbedFile 复制嵌入的文件到目标位置
func copyEmbedFile(embedFS embed.FS, embedFile, destinationFile string) error {
	// 打开嵌入的文件
	file, err := embedFS.Open(embedFile)
	if err != nil {
		return fmt.Errorf("failed to open embedded file %s: %w", embedFile, err)
	}
	defer file.Close()

	// 创建目标文件
	out, err := os.Create(destinationFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destinationFile, err)
	}
	defer out.Close()

	// 复制文件内容
	_, err = io.Copy(out, file)
	if err != nil {
		return fmt.Errorf("failed to copy content to destination file %s: %w", destinationFile, err)
	}

	return nil
}
