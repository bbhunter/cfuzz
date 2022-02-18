package fuzz

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	WordlistFilename string
	Keyword          string
	Command          string
	RoutineDelay     int64
	Shell            string
	Timeout          int64
	DisplayModes     []DisplayMode
	Filters          []Filter
	// OutputFilters    []display.OutputFilter
	// TimeFilters      []TimeFilter
	// CodeFilters      []CodeFilter
}

var usage = `Usage of cfuzz: cfuzz [flags values] [command] or cfuzz [flags values] [command] with CFUZZ_CMD environment variable set
Fuzz command line execution and filter results
  -w, --wordlist     wordlist used by fuzzer
  -f, --filter	     filter type to sort execution results (among 'output','time','return')
  -d, --delay        delay in ms between each thread launching. A thread execute the command. (default: 0)
  -k, --keyword      keyword use to determine which zone to fuzz (default: FUZZ)
  -s, --shell        shell to use for execution (default: /bin/bash)
  -to, --timeout      command execution timeout in s. After reaching it the command is killed. (default: 30)
  -h, --help         prints help information 
`

// NewConfig create Config instance
func NewConfig() Config {
	// default value
	config := Config{Keyword: "FUZZ"} //FilterType is already by default Output but to keep it in mind

	//flag wordlist
	flag.StringVar(&config.WordlistFilename, "wordlist", "", "wordlist used by fuzzer")
	flag.StringVar(&config.WordlistFilename, "w", "", "wordlist used by fuzzer")

	//flag keyword
	flag.StringVar(&config.Keyword, "keyword", "FUZZ", "keyword use to determine which zone to fuzz")
	flag.StringVar(&config.Keyword, "k", "FUZZ", "keyword use to determine which zone to fuzz")

	//flag shell
	flag.StringVar(&config.Shell, "shell", "/bin/bash", "shell to use for execution")
	flag.StringVar(&config.Shell, "s", "/bin/bash", "shell to use for execution")

	//flag RoutineDelay
	flag.Int64Var(&config.RoutineDelay, "d", 0, "delay in ms between each thread launching. A thread execute the command. (default: 0)")
	flag.Int64Var(&config.RoutineDelay, "delay", 0, "delay in ms between each thread launching. A thread execute the command. (default: 0)")

	//flag timeout
	flag.Int64Var(&config.Timeout, "to", 30, "Command execution timeout in s. After reaching it the command is killed. (default: 30)")
	flag.Int64Var(&config.Timeout, "timeout", 30, "Command execution timeout in s. After reaching it the command is killed. (default: 30)")

	// display mode
	var stdoutDisplay bool
	flag.BoolVar(&stdoutDisplay, "oc", false, "display command execution  number of characters in stdout.")
	flag.BoolVar(&stdoutDisplay, "stdout-characters", false, "display execution command number of characters in stdout.")

	var stderrDisplay bool
	flag.BoolVar(&stderrDisplay, "ec", false, "display command execution  number of characters in stderr.")
	flag.BoolVar(&stderrDisplay, "stderr-characters", false, "display execution command number of characters in stderr.")

	var timeDisplay bool
	flag.BoolVar(&timeDisplay, "t", false, "display command execution  time.")
	flag.BoolVar(&timeDisplay, "time", false, "display command execution time.")

	var codeDisplay bool
	flag.BoolVar(&codeDisplay, "c", false, "display command execution exit code.")
	flag.BoolVar(&codeDisplay, "code", false, "display command execution exit code.")

	// filters
	parseFilters(&config)

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	// command
	if cmdEnv := os.Getenv("CFUZZ_CMD"); cmdEnv != "" {
		config.Command = cmdEnv
	} else if flag.NArg() > 0 {
		cmdArg := strings.Join(flag.Args(), " ")
		config.Command = cmdArg
	}

	// parse display mode
	config.DisplayModes = parseDisplayMode(stdoutDisplay, stderrDisplay, timeDisplay, codeDisplay)

	return config
}

// CheckConfig: assert that all required fields are present in config, and are adequate to cfuzz run
func (c *Config) CheckConfig() error {
	// check field
	if c.WordlistFilename == "" {
		return errors.New("No wordlist provided. Please indicate a wordlist to use for fuzzing (-w,--wordlist)")
	}

	if c.Keyword == "" {
		return errors.New("Fuzzing Keyword can't be empty string")
	}
	if c.Command == "" {
		return errors.New("No command provided. Please indicate it using environment variable CFUZZ_CMD or cfuzz [flag:value] [command]")
	}

	// check field consistency
	if !strings.Contains(c.Command, c.Keyword) {
		return errors.New("Fuzzing keyword has not been found in command. keyword:" + c.Keyword + " command:" + c.Command)
	}

	return nil
}

// parseDisplayMode: parse display mode chosen with flag and return array of these display mode interface. If none, default is stdout characters display mode
func parseDisplayMode(stdout bool, stderr bool, time bool, code bool) (modes []DisplayMode) {
	if stdout {
		modes = append(modes, StdoutDisplay{})
	}
	if stderr {
		modes = append(modes, StderrDisplay{})
	}
	if time {
		modes = append(modes, TimeDisplay{})
	}
	if code {
		modes = append(modes, CodeDisplay{})
	}

	//default, if none
	if len(modes) < 0 {
		stdoutDisplay := StdoutDisplay{}
		modes = []DisplayMode{stdoutDisplay}
	}
	return modes
}

// parseFilters: parse all flags and determine the filters, add them in config struct
func parseFilters(config *Config) {
	// stdout filters
	flag.Func("omax", "filter to display only results with less than n characters", func(max string) error {
		n, err := strconv.Atoi(max)
		if err != nil {
			return err
		}
		filter := StdoutMaxFilter{Max: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	flag.Func("omin", "filter to display only results with more than n characters", func(min string) error {
		n, err := strconv.Atoi(min)
		if err != nil {
			return err
		}
		filter := StdoutMinFilter{Min: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	flag.Func("oeq", "filter to display only results with exactly n characters", func(eq string) error {
		n, err := strconv.Atoi(eq)
		if err != nil {
			return err
		}
		filter := StdoutEqFilter{Eq: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	// stderr filters
	flag.Func("emax", "filter to display only results with less than n characters", func(max string) error {
		n, err := strconv.Atoi(max)
		if err != nil {
			return err
		}
		filter := StderrMaxFilter{Max: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	flag.Func("emin", "filter to display only results with more than n characters", func(min string) error {
		n, err := strconv.Atoi(min)
		if err != nil {
			return err
		}
		filter := StderrMinFilter{Min: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	flag.Func("eeq", "filter to display only results with exactly n characters", func(eq string) error {
		n, err := strconv.Atoi(eq)
		if err != nil {
			return err
		}
		filter := StderrEqFilter{Eq: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	// time filters
	flag.Func("tmax", "filter to display only results with a time lesser than n seconds", func(max string) error {
		n, err := strconv.Atoi(max)
		if err != nil {
			return err
		}
		filter := TimeMaxFilter{Max: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	flag.Func("tmin", "filter to display only results with a time greater than n seconds", func(min string) error {
		n, err := strconv.Atoi(min)
		if err != nil {
			return err
		}
		filter := TimeMinFilter{Min: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	flag.Func("teq", "filter to  display only results with a time equal to n seconds", func(eq string) error {
		n, err := strconv.Atoi(eq)
		if err != nil {
			return err
		}
		filter := TimeEqFilter{Eq: n}
		config.Filters = append(config.Filters, filter)
		return nil
	})

	// code filters
	var success, failure bool
	flag.BoolVar(&success, "success", false, "filter to display only command with exit code 0.")
	flag.BoolVar(&failure, "failure", false, "filter to display only command with a non-zero exit .")
	if success {
		filter := CodeSuccessFilter{Zero: true}
		config.Filters = append(config.Filters, filter)
	}
	if failure {
		filter := CodeSuccessFilter{Zero: false}
		config.Filters = append(config.Filters, filter)
	}
}