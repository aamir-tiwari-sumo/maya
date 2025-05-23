/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package padd

import (
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	"github.com/aamir-tiwari-sumo/maya/pkg/zfs/cmd/v1alpha1/bin"
	"github.com/pkg/errors"
)

const (
	// Operation defines type of zfs operation
	Operation = "replace"
)

//PoolDiskReplace defines structure for pool 'Disk Replace' operation
type PoolDiskReplace struct {
	// Vdev to be replaced
	OldVdev string

	// New Vdev
	NewVdev string

	// property list
	Property []string

	// name of pool
	Pool string

	// Forcefully .. with -f
	Forcefully bool

	// command string
	Command string

	// checks is list of predicate function used for validating object
	checks []PredicateFunc

	// error
	err error
}

// NewPoolDiskReplace returns new instance of object PoolDiskReplace
func NewPoolDiskReplace() *PoolDiskReplace {
	return &PoolDiskReplace{}
}

// WithCheck add given check to checks list
func (p *PoolDiskReplace) WithCheck(check ...PredicateFunc) *PoolDiskReplace {
	p.checks = append(p.checks, check...)
	return p
}

// WithOldVdev method fills the OldVdev field of PoolDiskReplace object.
func (p *PoolDiskReplace) WithOldVdev(OldVdev string) *PoolDiskReplace {
	p.OldVdev = OldVdev
	return p
}

// WithNewVdev method fills the NewVdev field of PoolDiskReplace object.
func (p *PoolDiskReplace) WithNewVdev(NewVdev string) *PoolDiskReplace {
	p.NewVdev = NewVdev
	return p
}

// WithProperty method fills the Property field of PoolDiskReplace object.
func (p *PoolDiskReplace) WithProperty(key, value string) *PoolDiskReplace {
	p.Property = append(p.Property, fmt.Sprintf("%s=%s", key, value))
	return p
}

// WithForcefully method fills the Forcefully field of PoolDiskReplace object.
func (p *PoolDiskReplace) WithForcefully(Forcefully bool) *PoolDiskReplace {
	p.Forcefully = Forcefully
	return p
}

// WithPool method fills the Pool field of PoolDiskReplace object.
func (p *PoolDiskReplace) WithPool(Pool string) *PoolDiskReplace {
	p.Pool = Pool
	return p
}

// WithCommand method fills the Command field of PoolDiskReplace object.
func (p *PoolDiskReplace) WithCommand(Command string) *PoolDiskReplace {
	p.Command = Command
	return p
}

// Validate is to validate generated PoolDiskReplace object by builder
func (p *PoolDiskReplace) Validate() *PoolDiskReplace {
	for _, check := range p.checks {
		if !check(p) {
			p.err = errors.Wrapf(p.err, "validation failed {%v}", runtime.FuncForPC(reflect.ValueOf(check).Pointer()).Name())
		}
	}
	return p
}

// Execute is to execute generated PoolDiskReplace object
func (p *PoolDiskReplace) Execute() ([]byte, error) {
	p, err := p.Build()
	if err != nil {
		return nil, err
	}
	// execute command here
	// #nosec
	return exec.Command(bin.BASH, "-c", p.Command).CombinedOutput()
}

// Build returns the PoolDiskReplace object generated by builder
func (p *PoolDiskReplace) Build() (*PoolDiskReplace, error) {
	var c strings.Builder
	p = p.Validate()
	p.appendCommand(&c, bin.ZPOOL)
	p.appendCommand(&c, fmt.Sprintf(" %s ", Operation))

	if IsForcefullySet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" -f "))
	}

	if IsPropertySet()(p) {
		for _, v := range p.Property {
			p.appendCommand(&c, fmt.Sprintf(" -o %s ", v))
		}
	}

	p.appendCommand(&c, p.Pool)
	p.appendCommand(&c, fmt.Sprintf(" %s ", p.OldVdev))
	p.appendCommand(&c, fmt.Sprintf(" %s ", p.NewVdev))

	p.Command = c.String()
	return p, p.err
}

// appendCommand append string to given string builder
func (p *PoolDiskReplace) appendCommand(c *strings.Builder, cmd string) {
	_, err := c.WriteString(cmd)
	if err != nil {
		p.err = errors.Wrapf(p.err, "Failed to append cmd{%s} : %s", cmd, err.Error())
	}
}
