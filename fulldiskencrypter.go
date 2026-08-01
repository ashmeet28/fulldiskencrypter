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

		buf, err := os.ReadFile(inFilePath)
		if err != nil {
			log.Fatal(err)
		}

		if (len(buf) % encCypherBlockSize) != 0 {
			log.Fatal(errors.New("invalid input file size"))
		}

		for i := 0; i < len(buf); i += encCypherBlockSize {
			encHash := sha256.Sum256(binary.LittleEndian.AppendUint64(encKey, encCounter))
			encCounter++
			for j := range encCypherBlockSize {
				buf[i+j] = buf[i+j] ^ encHash[j]
			}
		}
		if err := os.WriteFile(outFilePath, buf, 0600); err != nil {
			log.Fatal(err)
		}

	case "genrate_bash_file":

		diskSize, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}

		diskFilePath := os.Args[3]
		tmpfsDirPath := os.Args[4]
		encKeyFilePath := tmpfsDirPath + "key"
		outFilePath := tmpfsDirPath + "run"

		if (diskSize % 4096) != 0 {
			log.Fatal(errors.New("invalid disk size"))
		}

		ioBlockSize := 67108864

		outFileBuf := make([]byte, 0)

		appendLineToBuf := func(s string) {
			outFileBuf = append(outFileBuf, []byte(s)...)
			outFileBuf = append(outFileBuf, 0x0a)
		}

		appendLineToBuf("#!/usr/bin/env bash")

		for i := 0; i < int(diskSize); i += ioBlockSize {

			appendLineToBuf("dd if=" + diskFilePath +
				" of=" + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) +
				" bs=" + strconv.Itoa(ioBlockSize) +
				" skip=" + strconv.Itoa(i/ioBlockSize) +
				" count=1" +
				" status=none")

			appendLineToBuf("[ $? -ne 0 ] && exit")

			appendLineToBuf(tmpfsDirPath + "fulldiskencrypter encrypt_file " +
				strconv.Itoa(i/encCypherBlockSize) +
				" " + encKeyFilePath +
				" " + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) +
				" " + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) + "_encrypted")

			appendLineToBuf("[ $? -ne 0 ] && exit")

			appendLineToBuf("dd if=" + tmpfsDirPath + strconv.Itoa(i/ioBlockSize) + "_encrypted" +
				" of=" + diskFilePath +
				" bs=" + strconv.Itoa(ioBlockSize) +
				" seek=" + strconv.Itoa(i/ioBlockSize) +
				" iflag=fullblock" +
				" conv=notrunc" +
				" status=none")

			appendLineToBuf("[ $? -ne 0 ] && exit")

			appendLineToBuf("sync")

			appendLineToBuf("[ $? -ne 0 ] && exit")

			if (i / ioBlockSize) >= 2 {
				appendLineToBuf("rm " + tmpfsDirPath + strconv.Itoa((i/ioBlockSize)-3) +
					" " + tmpfsDirPath + strconv.Itoa((i/ioBlockSize)-3) + "_encrypted")

				appendLineToBuf("[ $? -ne 0 ] && exit")
			}

		}

		if err := os.WriteFile(outFilePath, outFileBuf, 0600); err != nil {
			log.Fatal(err)
		}

	default:

		log.Fatal(errors.New("invalid arguments"))

	}
}
