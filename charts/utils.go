// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package charts

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func newInstallCommand() *exec.Cmd {
	return exec.Command("helm", "install")
}

func newUninstallCommand() *exec.Cmd {
	return exec.Command("helm", "delete", "--purge")
}

func installRelease(name string, args []string, path string, debug bool) error {
	cmd := newInstallCommand()

	if name != "" {
		cmd.Args = append(cmd.Args, "--name")
		cmd.Args = append(cmd.Args, name)
	}

	if len(args) > 0 {
		cmd.Args = append(cmd.Args, "--set")
		cmd.Args = append(cmd.Args, strings.Join(args, ","))
	}

	cmd.Args = append(cmd.Args, path)

	if debug {
		cmd.Args = append(cmd.Args, "--dry-run")
		cmd.Args = append(cmd.Args, "--debug")
	} else {
		cmd.Args = append(cmd.Args, "--wait")
		cmd.Args = append(cmd.Args, "--timeout")
		cmd.Args = append(cmd.Args, "600")
	}

	if debug {
		fmt.Println(cmd.Args)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err, string(output))
		return err
	}

	fmt.Println(string(output))
	return nil
}

func uninstallRelease(release string) error {
	cmd := newUninstallCommand()

	if release != "" {
		cmd.Args = append(cmd.Args, release)
	} else {
		return errors.New("Unknown release name")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(string(output))
	return nil
}

func ListCharts() {
	cmd := exec.Command("helm", "list")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))
}
