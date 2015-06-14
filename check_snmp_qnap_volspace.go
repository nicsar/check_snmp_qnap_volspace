/*
check_snmp_qnap_volspace - plugin to check free space on Qnap volume.

Version: 1.0
Author: Nicola Sarobba <nicola.sarobba@gmail.com>
Date: 2015-02-16
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nicsar/nagutils"
)

// snmpwalk command.
const snmpwalk = "/usr/bin/snmpwalk"

// Default timeout for SNMP.
const timeoutdfl = 10

// SNMP OID Syste Volume Table.
const SystemVolumeTableOID = "1.3.6.1.4.1.24681.1.2.17"

// SNMP OID Index of System Volume Table.
const SysVolumeIndexOID = ".1.3.6.1.4.1.24681.1.2.17.1.1"

// SystemVolumeEntry summarize an entry of the System volume table.
type SystemVolumeEntry struct {
	SysVolumeDescr     string
	SysVolumeTotalSize string
	SysVolumeFreeSize  string
	SysVolumeStatus    string // abnormal, invalid, unmounted, synchronyzing
}

// NormalizeUnit assures that we're using the same units in TotalSize and in FreeSize.
func (s *SystemVolumeEntry) NormalizeUnit() {
	var tszunit string
	if s.SysVolumeStatus == "Ready" {
		splittedT := strings.Split(s.SysVolumeTotalSize, " ")
		splittedF := strings.Split(s.SysVolumeFreeSize, " ")
		tszunit = splittedT[1]
		if splittedT[1] != splittedF[1] {
			result, _ := convertUnit(s.SysVolumeFreeSize, tszunit)
			s.SysVolumeFreeSize = result
		}
	}
}

// UoM returns the unit of measure used by the system volume entry.
func (s *SystemVolumeEntry) UoM() string {
	splittedT := strings.Split(s.SysVolumeTotalSize, " ")
	return splittedT[1]
}

// convertUnit converts vaule to 'dunit'. value is something like "8.44 TB".
func convertUnit(value string, dunit string) (string, error) {
	splitted := strings.Split(value, " ")
	svalue := splitted[0]
	sunit := splitted[1]
	if sunit == dunit { // Conversion not needed.
		return value, nil
	}
	var v float64
	var err error
	if v, err = strconv.ParseFloat(svalue, 64); err != nil {
		return "", fmt.Errorf("convertUnit: %s", err)
	}
	if sunit == "MB" && dunit == "GB" { // MB to GB
		return fmt.Sprintf("%f GB", v/1000), nil
	} else if sunit == "MB" && dunit == "TB" { // MB to TB
		return fmt.Sprintf("%f TB", v/1000000), nil
	} else if sunit == "GB" && dunit == "TB" { // GB to TB
		return fmt.Sprintf("%f TB", v/1000), nil
	} else if sunit == "GB" && dunit == "MB" { // GB to MB
		return fmt.Sprintf("%f MB", v*1000), nil
	} else if sunit == "TB" && dunit == "GB" { // TB to GB
		return fmt.Sprintf("%f GB", v*1000), nil
	} else if sunit == "TB" && dunit == "MB" { // TB to MB
		return fmt.Sprintf("%f MB", v*1000000), nil
	}
	return "", fmt.Errorf("Unknown unit: %s", sunit)
}

// usage prints the command usage and flag defaults.
var usage = func() {
	var basename string = nagutils.Basename(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s -H <host> [-C <snmp_community>] [-p <port>] [-t <timeout>]\n", basename)
	flag.PrintDefaults()
}

func main() {
	optHost := flag.String("H", "", "name or IP address of host to check")
	optCommty := flag.String("C", "public", "community name for the host's SNMP agent")
	optPort := flag.String("p", "161", "SNMP port")
	optW := flag.Int("w", 80, "percent of space volume used to generate WARNING state")
	optC := flag.Int("c", 90, "percent of space volume used to generate CRITICAL state")
	optPerf := flag.Bool("f", false, "perfparse compatible output")
	optTimeout := flag.Int("t", timeoutdfl, "timeout for SNMP in seconds")
	flag.Parse()
	if flag.NFlag() < 3 {
		usage()
		os.Exit(nagutils.Errors["UNKNOWN"])
	}
	if *optHost == "" {
		usage()
		os.Exit(nagutils.Errors["UNKNOWN"])
	}
	if *optTimeout <= 0 {
		*optTimeout = timeoutdfl
	}
	argTimeout := "-t" + strconv.Itoa(*optTimeout)
	var argHost string
	if *optPort != "161" {
		argHost = *optHost + ":" + *optPort
	} else {
		argHost = *optHost
	}

	cmd := exec.Command(snmpwalk, "-v2c", "-c", *optCommty, "-r0", "-On", argTimeout, argHost, SystemVolumeTableOID)
	snmpOut, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s", snmpOut)
		nagutils.NagiosExit(nagutils.Errors["UNKNOWN"], err.Error())
	}
	lines := strings.Split(fmt.Sprintf("%s", snmpOut), "\n")

	// Get the number of system volume entries.
	nentries := tableEntries(lines)
	var sysvolentries []SystemVolumeEntry
	if sysvolentries, err = extractData(lines, nentries); err != nil {
		_ = sysvolentries
		nagutils.NagiosExit(nagutils.Errors["UNKNOWN"], err.Error())
	}
	var totsz, freesz float64
	var exitVal nagutils.ExitVal
	exitVal.Status = nagutils.Errors["OK"]
	outMsg := ""
	outPerf := ""
	for i := 0; i < nentries; i++ {
		// If volume status is different from "Ready", it could be impossible to get size information.
		if sysvolentries[i].SysVolumeStatus != "Ready" {
			exitVal.Status = nagutils.Errors["CRITICAL"]
			outMsg = sysvolentries[i].SysVolumeDescr + " status " + sysvolentries[i].SysVolumeStatus
			exitVal.Problems = append(exitVal.Problems, outMsg)

		} else {
			// Let's assure that SysVolumeTotalSize and SysVolumeFreeSize are of same unit.
			sysvolentries[i].NormalizeUnit()
			if totsz, err = size2Float(sysvolentries[i].SysVolumeTotalSize); err != nil {
				// Exit!
				nagutils.NagiosExit(nagutils.Errors["UNKNOWN"], err.Error())
			}
			if freesz, err = size2Float(sysvolentries[i].SysVolumeFreeSize); err != nil {
				// Exit!
				nagutils.NagiosExit(nagutils.Errors["UNKNOWN"], err.Error())
			}
			used := int(nagutils.Round(usedPercent(totsz, freesz)))
			if used > *optW && used < *optC { // WARNING
				if exitVal.Status != nagutils.Errors["CRITICAL"] {
					exitVal.Status = nagutils.Errors["WARNING"]
				}
				outMsg = sysvolentries[i].SysVolumeDescr + " above warning threshold"
				outPerf = fmt.Sprintf("'%s'=%.2f%s;%v;%v;0;%.2f",
					sysvolentries[i].SysVolumeDescr,
					nagutils.RoundPlus(totsz-freesz, 2),
					sysvolentries[i].UoM(),
					int(nagutils.Round(percent(totsz, *optW))),
					int(nagutils.Round(percent(totsz, *optC))),
					totsz,
				)
				exitVal.Problems = append(exitVal.Problems, outMsg)
				exitVal.PerfData = append(exitVal.PerfData, outPerf)

			} else if used > *optC { // CRITICAL
				exitVal.Status = nagutils.Errors["CRITICAL"]
				if i != 0 {
					outPerf += " "
				}
				outMsg = sysvolentries[i].SysVolumeDescr + " above critical threshold"
				outPerf = fmt.Sprintf("'%s'=%.2f%s;%v;%v;0;%.2f",
					sysvolentries[i].SysVolumeDescr,
					nagutils.RoundPlus(totsz-freesz, 2),
					sysvolentries[i].UoM(),
					int(nagutils.Round(percent(totsz, *optW))),
					int(nagutils.Round(percent(totsz, *optC))),
					totsz,
				)
				exitVal.Problems = append(exitVal.Problems, outMsg)
				exitVal.PerfData = append(exitVal.PerfData, outPerf)
			} else { // System Volume entry OK.
				outPerf = fmt.Sprintf("'%s'=%.2f%s;%v;%v;0;%.2f",
					sysvolentries[i].SysVolumeDescr,
					nagutils.RoundPlus(totsz-freesz, 2),
					sysvolentries[i].UoM(),
					int(nagutils.Round(percent(totsz, *optW))),
					int(nagutils.Round(percent(totsz, *optC))),
					totsz,
				)
				exitVal.PerfData = append(exitVal.PerfData, outPerf)
			}
		}

	} // End for.
	// OK status.
	if exitVal.Status == nagutils.Errors["OK"] {
		msg := "volumes free space Ok - volumes status Ok"
		if *optPerf == true {
			msg += exitVal.PerfData2Str()
		}
		nagutils.NagiosExit(exitVal.Status, msg)
	}
	// WARNING or CRITICAL status.
	if exitVal.Status == nagutils.Errors["WARNING"] || exitVal.Status == nagutils.Errors["CRITICAL"] {
		msg := exitVal.Problems2Str()
		if *optPerf == true {
			msg += exitVal.PerfData2Str()
		}
		nagutils.NagiosExit(exitVal.Status, msg)
	}
	// We should never get here.
	nagutils.NagiosExit(nagutils.Errors["UNKNOWN"], "maybe a plugin bug")
}

// ** Functions **

// tableEntries returns the number of entries in SystemVolumeTable table.
func tableEntries(lines []string) int {
	entries := 0
	for _, line := range lines {
		// .1.3.6.1.4.1.24681.1.2.17.1.1.1 = INTEGER: 1
		splitted := strings.Split(line, "=")
		if strings.Contains(splitted[0], SysVolumeIndexOID) {
			entries++
		} else {
			break
		}
	}
	return entries
}

// extractData returns a slice of all the system volume entries found.
// 'lines' are the result of snmpwalk.
// 'nentries' are the number of entries in Sysvolume table.
func extractData(lines []string, nentries int) ([]SystemVolumeEntry, error) {

	var sysvolentries []SystemVolumeEntry

	descr := lines[nentries : nentries*2]
	totalsz := lines[nentries*3 : (nentries*3)+nentries]
	freesz := lines[nentries*4 : (nentries*4)+nentries]
	status := lines[nentries*5 : (nentries*5)+nentries]

	var out1, out2, out3, out4 string
	var err error
	for i := 0; i < nentries; i++ {
		if out1, err = extractDataValue(descr[i]); err != nil {
			return nil, err
		}
		if out2, err = extractDataValue(totalsz[i]); err != nil {
			return nil, err
		}
		if out3, err = extractDataValue(freesz[i]); err != nil {
			return nil, err
		}
		if out4, err = extractDataValue(status[i]); err != nil {
			return nil, err
		}
		sysvolentry := SystemVolumeEntry{out1, out2, out3, out4}
		sysvolentries = append(sysvolentries, sysvolentry)
	}
	return sysvolentries, nil
}

// extractDataValue returns a string containing the string value of the snmp oid line
// stripped of '"'.
func extractDataValue(snmpln string) (string, error) {

	splitted := strings.Split(snmpln, "= STRING: ")
	if len(splitted) != 2 {
		return "", fmt.Errorf("SNMP line: bad format of: %s", snmpln)
	}
	s := strings.Trim(splitted[1], "\"")
	return s, nil
}

// size2float convert the string "10.77 TB" in float64 10.77.
func size2Float(size string) (float64, error) {
	splitted := strings.Split(size, " ")
	f, err := strconv.ParseFloat(splitted[0], 64)
	if err != nil {
		return -1, fmt.Errorf("Bad size value: %s", size)
	}
	return f, nil

}

// usedPercent returns a float representing the percentage of used space.
func usedPercent(tot float64, free float64) float64 {
	used := tot - free
	percent := (used * 100) / tot
	return nagutils.RoundPlus(percent, 2)
}

// percent returns percentual value of 'n'.
func percent(n float64, prc int) float64 {
	return (n / 100.0) * float64(prc)
}
