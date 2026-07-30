package main

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"log"
	"os"
	"strconv"
)

func main() {
	opToExec := os.Args[1]
	encCypherBlockSize := 32

	switch opToExec {

	case "encrypt_file":
		encCounter, err := strconv.ParseUint(os.Args[2], 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		encKeyFilePath := os.Args[3]
		inFilePath := os.Args[4]
		outFilePath := os.Args[5]

		encKey, err := os.ReadFile(encKeyFilePath)
		if err != nil {
			log.Fatal(err)
		}

		curInBuf, err := os.ReadFile(inFilePath)
		if err != nil {
			log.Fatal(err)
		}

		if (len(curInBuf) % encCypherBlockSize) != 0 {
			log.Fatal(errors.New("invalid input file size"))
		}

		curOutBuf := make([]byte, len(curInBuf))
		for i := 0; i < len(curInBuf); i += encCypherBlockSize {
			encHash := sha256.Sum256(binary.LittleEndian.AppendUint64(encKey, encCounter))
			encCounter++
			for j := range encCypherBlockSize {
				curOutBuf[i+j] = curInBuf[i+j] ^ encHash[j]
			}
		}
		if err := os.WriteFile(outFilePath, curOutBuf, 0600); err != nil {
			log.Fatal(err)
		}

	case "genrate_bash_file":
		diskSize, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		diskFilePath := os.Args[3]
		encKeyFilePath := os.Args[4]
		tmpfsDirPath := os.Args[5]

		if (diskSize % 4096) != 0 {
			log.Fatal(errors.New("invalid disk size"))
		}
		maxIoBlockSize := 67108864
		ioBlockSize := maxIoBlockSize
		for (diskSize % ioBlockSize) != 0 {
			ioBlockSize = ioBlockSize / 2
		}

		for i := 0; i < int(diskSize); i += ioBlockSize {
			stdoutBuf := make([]byte, 0)

			appendLineToStdoutBuf := func(s string) {
				stdoutBuf = append(stdoutBuf, []byte(s)...)
				stdoutBuf = append(stdoutBuf, 0x0a)
			}

			if i == 0 {
				appendLineToStdoutBuf("#!/usr/bin/env bash")
			}

			appendLineToStdoutBuf("dd if=" + diskFilePath +
				" of=" + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) +
				" bs=" + strconv.Itoa(ioBlockSize) +
				" skip=" + strconv.Itoa(i/ioBlockSize) +
				" count=1" +
				" status=none")

			appendLineToStdoutBuf("[ $? -ne 0 ] && exit")

			appendLineToStdoutBuf(tmpfsDirPath + "fulldiskencrypter encrypt_file " + strconv.Itoa(i/encCypherBlockSize) +
				" " + encKeyFilePath +
				" " + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) +
				" " + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) + "_encrypted")

			appendLineToStdoutBuf("[ $? -ne 0 ] && exit")

			appendLineToStdoutBuf("dd if=" + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) + "_encrypted" +
				" of=" + diskFilePath +
				" bs=" + strconv.Itoa(ioBlockSize) +
				" seek=" + strconv.Itoa(i/ioBlockSize) +
				" iflag=fullblock" +
				" conv=notrunc" +
				" status=none")

			appendLineToStdoutBuf("[ $? -ne 0 ] && exit")

			appendLineToStdoutBuf("sync")

			appendLineToStdoutBuf("[ $? -ne 0 ] && exit")

			if (i != 0) && (i/ioBlockSize) >= 3 {
				appendLineToStdoutBuf("rm " + tmpfsDirPath + strconv.Itoa((i/ioBlockSize)-3) +
					" " + tmpfsDirPath + strconv.Itoa((i/ioBlockSize)-3) + "_encrypted")

				appendLineToStdoutBuf("[ $? -ne 0 ] && exit")
			}

			if _, err := os.Stdout.Write(stdoutBuf); err != nil {
				log.Fatal(err)
			}
		}

	default:
		log.Fatal(errors.New("invalid arguments"))
	}
}
