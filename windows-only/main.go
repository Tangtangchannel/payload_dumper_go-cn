package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

func extractPayloadBin(filename string) string {
	zipReader, err := zip.OpenReader(filename)
	if err != nil {
		log.Fatalf("此zip不可用（通常在zip内含有一个payload.bin）: %s\n", filename)
	}
	defer zipReader.Close()

	for _, file := range zipReader.Reader.File {
		if file.Name == "payload.bin" && file.UncompressedSize64 > 0 {
			zippedFile, err := file.Open()
			if err != nil {
				log.Fatalf("此压缩文件无法被读取，是否可能损坏？: %s\n", file.Name)
			}

			tempfile, err := ioutil.TempFile(os.TempDir(), "payload_*.bin")
			if err != nil {
				log.Fatalf("无法创建临时文件，位于： %s\n", tempfile.Name())
			}
			defer tempfile.Close()

			_, err = io.Copy(tempfile, zippedFile)
			if err != nil {
				log.Fatal(err)
			}

			return tempfile.Name()
		}
	}

	return ""
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	list := flag.Bool("l", false, "列出payload中的分区")
	partitions := flag.String("p", "", "只转储选定分区")
	targetDir := flag.String("d", "all", "转储所有分区")

	fmt.Println(targetDir)
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
	}
	filename := flag.Arg(0)

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatalf("指定的文件不存在！: %s\n", filename)
	}

	payloadBin := filename
	if strings.HasSuffix(filename, ".zip") {
		fmt.Println("正在从zip归档文件中提取payload.bin，请等待...")
		payloadBin = extractPayloadBin(filename)
		if payloadBin == "" {
			log.Fatal("提取失败！（无法从归档文件中提取payload.bin）")
		} else {
			defer os.Remove(payloadBin)
		}
	}
	fmt.Printf("payload.bin: %s\n", payloadBin)

	payload := NewPayload(payloadBin)
	if err := payload.Open(); err != nil {
		log.Fatal(err)
	}
	payload.Init()

	if *list {
		return
	}

	now := time.Now()
	targetDirectory := fmt.Sprintf("extracted_%d%02d%02d_%s", now.Year(), now.Month(), now.Day(), *targetDir)

	if *partitions != "" {
		if err := payload.ExtractSelected(targetDirectory, strings.Split(*partitions, ",")); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := payload.ExtractAll(targetDirectory); err != nil {
			log.Fatal(err)
		}
	}
}

func usage() {
	fmt.Print("趣味小知识：你可以直接放入zip文件，程序会自行解压！\n")
	fmt.Fprintf(os.Stderr, "用法: %s [参数] [文件]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}
