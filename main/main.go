package main

import (
	"flag"
	//"fmt"
	"log"
	"math/rand"

	"../pidiver"
	"../raspberry"
	"github.com/iotaledger/giota"
)

const APP_VERSION = "0.1"

// The flag package provides a default help printer via -h switch
var versionFlag *bool = flag.Bool("v", false, "Print the version number.")
var forceFlash *bool = flag.Bool("force-upload", false, "Force file upload to SPI flash")
var forceConfigure *bool = flag.Bool("force-configure", false, "Force to configure FPGA from SPI flash")
var configFile *string = flag.String("fpga-config", "output_file.rbf", "FPGA config file to upload to SPI flash")
var device *string = flag.String("device", "/dev/ttyACM0", "Device file for usb communication")

var useUSB *bool = flag.Bool("usbdiver", false, "Use USB instead of Pi-GPIO")

func main() {
	flag.Parse() // Scan the arguments list

	config := pidiver.PiDiverConfig{
		Device:         *device,
		ConfigFile:     *configFile,
		ForceFlash:     *forceFlash,
		ForceConfigure: *forceConfigure}

	var powFunc giota.PowFunc
	var err error
	if *useUSB {
		err = pidiver.InitUSBDiver(&config)
		powFunc = pidiver.PowUSBDiver
	} else {
		llStruct := raspberry.GetLowLevel()
		err = pidiver.InitPiDiver(&llStruct, &config)
		powFunc = pidiver.PowPiDiver
	}
	if err != nil {
		log.Fatal(err)
	}

	// test transaction data
	var transaction string = "999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999A9RGRKVGWMWMKOLVMDFWJUHNUNYWZTJADGGPZGXNLERLXYWJE9WQHWWBMCPZMVVMJUMWWBLZLNMLDCGDJ999999999999999999999999999999999999999999999999999999YGYQIVD99999999999999999999TXEFLKNPJRBYZPORHZU9CEMFIFVVQBUSTDGSJCZMBTZCDTTJVUFPTCCVHHORPMGCURKTH9VGJIXUQJVHK999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999999"
	mwm := 14
	randomTrytes := make([]rune, 256)
	for i := 0; i < 10000; i++ {
		for j := 0; j < 256; j++ {
			randomTrytes[j] = rune(pidiver.TRYTE_CHARS[rand.Intn(len(pidiver.TRYTE_CHARS))])
		}
		var ret giota.Trytes
		var err error

		ret, err = powFunc(giota.Trytes(string(randomTrytes)+transaction[256:]), mwm)
		if err != nil {
			log.Fatalf("Error: %g", err)
		}
		log.Printf("Nonce-Trytes: %s\n", ret)
		// verify result
		trytes := giota.Trytes(string(randomTrytes) + transaction[256:len(transaction)-27] + string(ret[0:27]))
		log.Printf("hash: %s\n\n", trytes.Hash())
		tritsHash := trytes.Hash().Trits()
		for i := 0; i < mwm; i++ {
			if tritsHash[len(tritsHash)-1-i] != 0 {
				log.Fatalf("verify error at %d!\n", i)
			}
		}
	}

}
