/*
Copyright (c) 2009-2010 Jim Teeuwen.

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following restrictions:

    1. The origin of this software must not be misrepresented; you must not
    claim that you wrote the original software. If you use this software
    in a product, an acknowledgment in the product documentation would be
    appreciated but is not required.

    2. Altered source versions must be plainly marked as such, and must not be
    misrepresented as being the original software.

    3. This notice may not be removed or altered from any source distribution.

*/
package optarg

import "fmt"
import "os"
import "strings"
import "strconv"

type Option struct {
	Name        string
	ShortName   string
	Description string
	defaultval  interface{}
	value       string
}

var (
	options     = make([]*Option, 0)
	Remainder   = make([]string, 0)
	ShortSwitch = "-"
	LongSwitch  = "--"
	UsageInfo   = fmt.Sprintf("Usage: %s [options]:", os.Args[0])
)

const headerName = "__hdr"

// Prints usage information in a neatly formatted overview.
func Usage() {
	offset := 0

	// Find the largest length of the option name list. Needed to align
	// the description blocks consistently.
	for _, v := range options {
		if v.ShortName == headerName {
			continue
		}

		str := fmt.Sprintf("%s%s, %s%s: ", LongSwitch, v.Name, ShortSwitch, v.ShortName)
		if len(str) > offset {
			offset = len(str)
		}
	}

	offset++ // add margin.

	fmt.Printf("%s\n", UsageInfo)

	for _, v := range options {
		if v.ShortName == headerName {
			fmt.Printf("\n[%s]\n", v.Description)
			continue
		}

		// Print namelist. right-align it based on the maximum width
		// found in previous loop.
		str := fmt.Sprintf("%s%s, %s%s: ", LongSwitch, v.Name, ShortSwitch, v.ShortName)
		format := fmt.Sprintf("%%%ds", offset)
		fmt.Printf(format, str)

		desc := v.Description
		// boolean flags need no 'default value' description. They are either
		// set or not.
		if _, ok := v.defaultval.(bool); !ok {
			if fmt.Sprintf("%v", v.defaultval) != "" {
				desc = fmt.Sprintf("%s (defaults to: %v)", desc, v.defaultval)
			}
		}

		// Format and print left-aligned, word-wrapped description with
		// a @margin left margin size using my super duper patented
		// multi-line string wrap routine (see string.go). Assume
		// maximum of 80 characters screen width. Which makes block
		// width equal to 80 - @offset. I would prefer to use
		// ALIGN_JUSTIFY for added sexy, but it looks a little funky for
		// short descriptions. So we'll stick with the establish left-
		// aligned text.
		lines := multilineWrap(desc, 80, offset, 0, ALIGN_LEFT)

		// First line needs to be appended to where we left off.
		fmt.Printf("%s\n", strings.TrimSpace(lines[0]))

		// Print the rest as-is (properly indented).
		for i := 1; i < len(lines); i++ {
			fmt.Printf("%s\n", lines[i])
		}
	}
}

// Parse os.Args using the previously added Options.
func Parse() <-chan *Option {
	c := make(chan *Option)
	Remainder = make([]string, 0)
	go processArgs(c)
	return c
}

func processArgs(c chan *Option) {
	var opt *Option
	for i, v := range os.Args {
		if i == 0 {
			continue
		} // skip app name

		if v = strings.TrimSpace(v); len(v) == 0 {
			continue
		}

		if len(v) >= 3 && v[0:2] == LongSwitch {
			v := strings.TrimSpace(v[2:len(v)])
			if len(v) == 0 {
				Remainder = append(Remainder, LongSwitch)
			} else {
				opt = findOption(v)
				if opt == nil {
					fmt.Fprintf(os.Stderr, "Unknown option '--%s' specified.\n", v)
					Usage()
					os.Exit(1)
				}

				_, ok := opt.defaultval.(bool)
				if ok {
					opt.value = "true"
					c <- opt
					opt = nil
				}
			}

		} else if len(v) >= 2 && v[0:1] == ShortSwitch {
			if v = strings.TrimSpace(v[1:len(v)]); len(v) == 0 {
				Remainder = append(Remainder, ShortSwitch)
			} else {
				for i, _ := range v {
					tok := v[i : i+1]
					opt = findOption(tok)
					if opt == nil {
						fmt.Fprintf(os.Stderr, "Unknown option '-%s' specified.\n", tok)
						Usage()
						os.Exit(1)
					}

					_, ok := opt.defaultval.(bool)
					if ok {
						opt.value = "true"
						c <- opt
						opt = nil
					}
				}
			}

		} else {
			if opt == nil {
				Remainder = append(Remainder, v)
			} else {
				opt.value = v
				c <- opt
				opt = nil
			}
		}
	}
	close(c)
}

// Adds a section header. Useful to separate options into discrete groups.
func Header(title string) {
	options = append(options, &Option{
		ShortName:   headerName,
		Description: title,
	})
}

// Add a new command line option to check for.
func Add(shortname, name, description string, defaultvalue interface{}) {
	options = append(options, &Option{
		ShortName:   shortname,
		Name:        name,
		Description: description,
		defaultval:  defaultvalue,
	})
}

func findOption(name string) *Option {
	for _, opt := range options {
		if opt.Name == name || opt.ShortName == name {
			return opt
		}
	}
	return nil
}

func (this *Option) String() string { return this.value }

func (this *Option) Bool() bool {
	if b, err := strconv.Atob(this.value); err == nil {
		return b
	}
	return false
}

func (this *Option) Int() int {
	if v, err := strconv.Atoi(this.value); err == nil {
		return v
	}
	return this.defaultval.(int)
}
func (this *Option) Int8() int8   { return int8(this.Int()) }
func (this *Option) Int16() int16 { return int16(this.Int()) }
func (this *Option) Int32() int32 { return int32(this.Int()) }
func (this *Option) Int64() int64 {
	if v, err := strconv.Atoi64(this.value); err == nil {
		return v
	}
	return this.defaultval.(int64)
}

func (this *Option) Uint() uint {
	if v, err := strconv.Atoui(this.value); err == nil {
		return v
	}
	return this.defaultval.(uint)
}
func (this *Option) Uint8() uint8   { return uint8(this.Int()) }
func (this *Option) Uint16() uint16 { return uint16(this.Int()) }
func (this *Option) Uint32() uint32 { return uint32(this.Int()) }
func (this *Option) Uint64() uint64 {
	if v, err := strconv.Atoui64(this.value); err == nil {
		return v
	}
	return this.defaultval.(uint64)
}

func (this *Option) Float() float {
	if v, err := strconv.Atof(this.value); err == nil {
		return v
	}
	return this.defaultval.(float)
}

func (this *Option) Float32() float32 {
	if v, err := strconv.Atof32(this.value); err == nil {
		return v
	}
	return this.defaultval.(float32)
}

func (this *Option) Float64() float64 {
	if v, err := strconv.Atof64(this.value); err == nil {
		return v
	}
	return this.defaultval.(float64)
}
