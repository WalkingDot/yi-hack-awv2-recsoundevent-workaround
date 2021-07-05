package main

import (
	"bufio"
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	flagrectime := flag.Duration("t", 10, "Recording duration in minutes (1-60, default: 10).\nExample: -t 2")
	flagdblvl := flag.Int("d", 0, "dB value to trigger recording (30-90, default: Cam dB setting).\nExample: -d 42")
	flag.Parse()

	if _, err := os.Stat("/tmp/sd/yi-hack/ipccmdcodes"); os.IsNotExist(err) {
		if err := os.Mkdir("/tmp/sd/yi-hack/ipccmdcodes", 0777); err != nil {
			os.Exit(1)
		}
		if err := os.WriteFile("/tmp/sd/yi-hack/ipccmdcodes/ipccmdrecon.bin", []byte{1, 0, 0, 0, 2, 0, 0, 0, 124, 0, 124, 0, 0, 0, 0, 0}, 0666); err != nil {
			os.Exit(2)
		}
		if err := os.WriteFile("/tmp/sd/yi-hack/ipccmdcodes/ipccmdrecoff.bin", []byte{1, 0, 0, 0, 2, 0, 0, 0, 125, 0, 125, 0, 0, 0, 0, 0}, 0666); err != nil {
			os.Exit(3)
		}
	}

	SetdBlvl(*flagdblvl)

	if err := KillbyName("ipc_multiplexer"); err != nil {
		os.Exit(4)
	}

	cmd := exec.Command("/tmp/sd/yi-hack/bin/ipc_multiplexer")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		os.Exit(5)
	}

	if err := cmd.Start(); err != nil {
		os.Exit(6)
	}

	var recording bool
	var recduration time.Time
	mpscan := bufio.NewScanner(stderr)
	for mpscan.Scan() {
		if recording {
			if mpscan.Text() == "04 00 00 00 02 00 00 00 04 60 04 60 00 00 00 00 " {
				recduration = recduration.Add(time.Minute * *flagrectime)
			} else if mpscan.Text() == "01 00 00 00 02 00 00 00 7d 00 7d 00 00 00 00 00 " && time.Until(recduration).Minutes() > 0 {
				if err := exec.Command("/tmp/sd/yi-hack/bin/ipc_cmd", "-f", "/tmp/sd/yi-hack/ipccmdcodes/ipccmdrecon.bin").Run(); err != nil {
					os.Exit(7)
				}
			} else if time.Until(recduration).Minutes() < 0 {
				if err := exec.Command("/tmp/sd/yi-hack/bin/ipc_cmd", "-f", "/tmp/sd/yi-hack/ipccmdcodes/ipccmdrecoff.bin").Run(); err != nil {
					os.Exit(8)
				}
				recording = false
			}
		} else {
			if mpscan.Text() == "04 00 00 00 02 00 00 00 04 60 04 60 00 00 00 00 " {
				if err := exec.Command("/tmp/sd/yi-hack/bin/ipc_cmd", "-f", "/tmp/sd/yi-hack/ipccmdcodes/ipccmdrecon.bin").Run(); err != nil {
					os.Exit(9)
				}
				recording = true
				recduration = time.Now().Add(time.Minute * *flagrectime)
			}
		}
	}

	if err := mpscan.Err(); err != nil {
		os.Exit(10)
	}

	if err := cmd.Wait(); err != nil {
		os.Exit(11)
	}
}

func KillbyName(s string) error {
	if pids, err := filepath.Glob("/proc/[0-9]*"); err != nil {
		return errors.New("kill: proc failed")
	} else {
		for _, pid := range pids {
			if process, _ := os.Readlink(pid + "/exe"); process == "/tmp/sd/yi-hack/bin/"+s {
				pidint, _ := strconv.Atoi(filepath.Base(pid))
				proc, _ := os.FindProcess(pidint)
				if err := proc.Kill(); err != nil {
					return errors.New("kill: failed")
				}
				for i := 0; i < 5; i++ {
					if _, err := os.Stat(pid); os.IsNotExist(err) {
						return nil
					}
					time.Sleep(500 * time.Millisecond)
				}
				return errors.New("kill: timeout")
			}
		}
		return errors.New("kill: process not found")
	}
}

func SetdBlvl(db int) error {
	if db > 29 && db < 91 {
		ipccmdhex := []byte{2, 0, 0, 0, 8, 0, 0, 0, 59, 16, 1, 0, 4, 0, 0, 0, 0, 0, 0, 0}
		ipccmdhex[16] = byte(db)
		if err := os.WriteFile("/tmp/sd/yi-hack/ipccmdcodes/ipccmddb.bin", ipccmdhex, 0666); err != nil {
			return errors.New("dblvl: write failed")
		}
		if err := exec.Command("/tmp/sd/yi-hack/bin/ipc_cmd", "-f", "/tmp/sd/yi-hack/ipccmdcodes/ipccmddb.bin").Run(); err != nil {
			return errors.New("dblvl: send dblvl failed")
		}
		return nil
	} else if db == 0 {
		return nil
	} else {
		return errors.New("dblvl: out of range")
	}
}
