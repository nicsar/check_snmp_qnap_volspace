package main

import (
	"fmt"
	"testing"
)

var oidLinesSamples = [][]string{
	{
		".1.3.6.1.4.1.24681.1.2.17.1.1.1 = INTEGER: 1",
		".1.3.6.1.4.1.24681.1.2.17.1.1.2 = INTEGER: 2",
		".1.3.6.1.4.1.24681.1.2.17.1.1.3 = INTEGER: 3",
		".1.3.6.1.4.1.24681.1.2.17.1.2.1 = STRING: \"[Volume Volume-1, Pool 1]\"",
		".1.3.6.1.4.1.24681.1.2.17.1.2.2 = STRING: \"[Volume Volume-2, Pool 2]\"",
		".1.3.6.1.4.1.24681.1.2.17.1.2.3 = STRING: \"[Single Disk Volume:  REXP#34 Drive: 6]\"",
		".1.3.6.1.4.1.24681.1.2.17.1.3.1 = STRING: \"EXT4\"",
		".1.3.6.1.4.1.24681.1.2.17.1.3.2 = STRING: \"EXT4\"",
		".1.3.6.1.4.1.24681.1.2.17.1.3.3 = STRING: \"EXT4\"",
		".1.3.6.1.4.1.24681.1.2.17.1.4.1 = STRING: \"10.77 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.4.2 = STRING: \"5.35 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.4.3 = STRING: \"916.39 GB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.5.1 = STRING: \"8.79 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.5.2 = STRING: \"3.79 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.5.3 = STRING: \"102.00 GB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.6.1 = STRING: \"Ready\"",
		".1.3.6.1.4.1.24681.1.2.17.1.6.2 = STRING: \"Ready\"",
		".1.3.6.1.4.1.24681.1.2.17.1.6.3 = STRING: \"Ready\""},

	{
		".1.3.6.1.4.1.24681.1.2.17.1.1.1 = INTEGER: 1",
		".1.3.6.1.4.1.24681.1.2.17.1.1.2 = INTEGER: 2",
		".1.3.6.1.4.1.24681.1.2.17.1.2.1 = STRING: \"[Volume Volume-1, Pool 1]\"",
		".1.3.6.1.4.1.24681.1.2.17.1.2.2 = STRING: \"[Volume Volume-2, Pool 2]\"",
		".1.3.6.1.4.1.24681.1.2.17.1.3.1 = STRING: \"EXT4\"",
		".1.3.6.1.4.1.24681.1.2.17.1.3.2 = STRING: \"EXT4\"",
		".1.3.6.1.4.1.24681.1.2.17.1.4.1 = STRING: \"10.77 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.4.2 = STRING: \"5.35 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.5.1 = STRING: \"8.79 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.5.2 = STRING: \"3.79 TB\"",
		".1.3.6.1.4.1.24681.1.2.17.1.6.1 = STRING: \"Ready\"",
		".1.3.6.1.4.1.24681.1.2.17.1.6.2 = STRING: \"Ready\""},
}

func TestConvertUnit(t *testing.T) {
	type testvalues struct {
		in      string
		unit    string
		correct string
		err     error
	}
	var cases = []testvalues{
		{"1500 MB", "GB", "1.500000 GB", nil},
		{"300 MB", "TB", "0.000300 TB", nil},
		{"228 GB", "TB", "0.228000 TB", nil},
		{"0.77 TB", "MB", "770000.000000 MB", nil},
		{"750 GB", "TB", "0.750000 TB", nil},
		{"890.5 GB", "MB", "890500.000000 MB", nil},
		{"150 TB", "GB", "150000.000000 GB", nil},
		{"150 TB", "MB", "150000000.000000 MB", nil},
	}
	for _, c := range cases {
		got, err := convertUnit(c.in, c.unit)
		if got != c.correct || err != c.err {
			t.Errorf("convertUnit(%q,%q) == %q,%q, correct %q,%q", c.in, c.unit, got, err, c.correct, c.err)
		}
	}
}
func TestConvertUnitErr(t *testing.T) {
	type testvalues struct {
		in      string
		unit    string
		correct string
	}
	var cases = []testvalues{
		{"150,80 TB", "GB", ""},
		{"150.80 TB", "KB", ""},
	}
	for _, c := range cases {
		got, err := convertUnit(c.in, c.unit)
		if err == nil || c.correct != "" {
			t.Errorf("convertUnit(%q,%q) expected null string and an error, got %q", c.in, got)
		}
	}
}

func TestNormalizeUnit(t *testing.T) {
	s1 := SystemVolumeEntry{"[Volume Volume-1, Pool 1]", "10.77 TB", "8.79 TB", "Ready"}
	s2 := SystemVolumeEntry{"[Volume Volume-1, Pool 1]", "10.77 TB", "388 MB", "Ready"}
	type testvalues struct {
		in      SystemVolumeEntry
		correct SystemVolumeEntry
	}
	var cases = []testvalues{
		{s1, SystemVolumeEntry{"[Volume Volume-1, Pool 1]", "10.77 TB", "8.79 TB", "Ready"}},
		{s2, SystemVolumeEntry{"[Volume Volume-1, Pool 1]", "10.77 TB", "0.000388 TB", "Ready"}},
	}
	for _, c := range cases {
		c.in.NormalizeUnit()
		if c.in.SysVolumeDescr != c.correct.SysVolumeDescr {
			t.Errorf("s.NormalizeUnit(), SysVolumeDescr modified!")
		}
		if c.in.SysVolumeTotalSize != c.correct.SysVolumeTotalSize {
			t.Errorf("s.NormalizeUnit(), SysVolumeTotalSize modified!")
		}
		if c.in.SysVolumeFreeSize != c.correct.SysVolumeFreeSize {
			t.Errorf("s.NormalizeUnit(), SysVolumeFreeSize != %q, got %q", c.correct.SysVolumeFreeSize, c.in.SysVolumeFreeSize)
		}
		if c.in.SysVolumeStatus != c.correct.SysVolumeStatus {
			t.Errorf("s.NormalizeUnit(), SysVolumeStatus modified!")
		}
	}
}

func TestUoM(t *testing.T) {
	s1 := SystemVolumeEntry{"[Volume Volume-1, Pool 1]", "10.77 TB", "8.79 TB", "Ready"}
	s2 := SystemVolumeEntry{"[Volume Volume-1, Pool 1]", "50.57 GB", "8.79 GB", "Ready"}
	type testvalues struct {
		in      SystemVolumeEntry
		correct string
	}
	var cases = []testvalues{
		{s1, "TB"},
		{s2, "GB"},
	}
	for _, c := range cases {
		got := c.in.UoM()
		if got != c.correct {
			t.Errorf("c.UoM(%q) == %q, correct %q", c.in, got, c.correct)
		}
	}
}

func TestTableEntries(t *testing.T) {
	type testvalues struct {
		in      []string
		correct int
	}
	var cases = []testvalues{
		{oidLinesSamples[0], 3},
		{oidLinesSamples[1], 2},
	}
	for _, c := range cases {
		got := tableEntries(c.in)
		if got != c.correct {
			t.Errorf("tableEntries(%q) == %q, correct %q", c.in, got, c.correct)
		}
	}
}

func TestExtractDataValue1(t *testing.T) {
	type testvalues struct {
		in      string
		correct string
		err     error
	}
	var cases = []testvalues{
		{".1.3.6.1.4.1.24681.1.2.17.1.2.1 = STRING: \"[Volume Volume-1, Pool 1]\"", "[Volume Volume-1, Pool 1]", nil},
		{".1.3.6.1.4.1.24681.1.2.17.1.4.1 = STRING: \"10.77 TB\"", "10.77 TB", nil},
		{".1.3.6.1.4.1.24681.1.2.17.1.5.1 = STRING: \"8.79 TB\"", "8.79 TB", nil},
		{".1.3.6.1.4.1.24681.1.2.17.1.6.1 = STRING: \"Ready\"", "Ready", nil},
	}
	for _, c := range cases {
		got, err := extractDataValue(c.in)
		if got != c.correct || err != c.err {
			t.Errorf("extractDataValue(%q) == %q, correct %q,%q", c.in, got, err, c.correct)
		}
	}
}
func TestExtractDataValue2(t *testing.T) {
	errcase := ".1.3.6.1.4.1.24681.1.2.17.1.1.1 = INTEGER: 1"
	got, e := extractDataValue(errcase)
	if e == nil {
		t.Errorf("extractDataValue(%q) expected an error, got %q", errcase, got)
	}
}

func TestExtractData(t *testing.T) {
	type testvalues struct {
		lines    []string
		nentries int
		correct  []SystemVolumeEntry
		err      error
	}
	sysvolentries1 := []SystemVolumeEntry{
		{"[Volume Volume-1, Pool 1]", "10.77 TB", "8.79 TB", "Ready"},
		{"[Volume Volume-2, Pool 2]", "5.35 TB", "3.79 TB", "Ready"},
		{"[Single Disk Volume:  REXP#34 Drive: 6]", "916.39 GB", "102.00 GB", "Ready"},
	}
	sysvolentries2 := []SystemVolumeEntry{
		{"[Volume Volume-1, Pool 1]", "10.77 TB", "8.79 TB", "Ready"},
		{"[Volume Volume-2, Pool 2]", "5.35 TB", "3.79 TB", "Ready"},
	}

	var cases = []testvalues{
		{oidLinesSamples[0], 3, sysvolentries1, nil},
		{oidLinesSamples[1], 2, sysvolentries2, nil},
	}
	for _, c := range cases {
		got, err := extractData(c.lines, c.nentries)
		strGot := fmt.Sprintf("%v", got)
		strCorrect := fmt.Sprintf("%v", c.correct)
		if strGot != strCorrect || err != c.err {
			t.Errorf("extractData(%q, %q) == %q, correct %q,%q", c.lines, c.nentries, got, c.correct, err)
		}
	}
}

func TestSize2Float(t *testing.T) {
	type testvalues struct {
		size    string
		correct float64
		err     error
	}
	var cases = []testvalues{
		{"10.77 TB", 10.77, nil},
		{"9.50 GB", 9.50, nil},
		{"300.45 MB", 300.45, nil},
		{"517.00", 517.00, nil},
	}
	var errcases = []testvalues{
		{"abc", -1, nil}, // err field is not really used here.
	}
	for _, c := range cases {
		got, err := size2Float(c.size)
		if got != c.correct || err != c.err {
			t.Errorf("size2Float(%q) == %q, correct %q,%q", c.size, got, c.correct, c.err)
		}
	}
	for _, c := range errcases {
		got, err := size2Float(c.size)
		if got != c.correct || err == nil {
			t.Errorf("size2Float(%q) == %q, correct %q,%q", c.size, got, c.correct, err)
		}
	}
}

func TestUsedPercent(t *testing.T) {
	type testvalues struct {
		tot     float64
		free    float64
		correct float64
	}
	var cases = []testvalues{
		{100.00, 20.00, 80.00},
		{310, 232.6, 24.97},
		{100.00, 0, 100.00},
	}
	for _, c := range cases {
		got := usedPercent(c.tot, c.free)
		if got != c.correct {
			t.Errorf("usedPercent(%q, %q) == %q, correct %q", c.tot, c.free, got, c.correct)
		}
	}
}
