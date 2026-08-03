package logs

const DefaultOutputFolder = "/tmp/logs/"
const DefaultOutputFile = "log.txt"

type Config struct {
	level   string      `toml:"level"`
	context LineContext `toml:"context"`
	output  Output      `toml:"output"`
}

type LineContext struct {
	format string            `toml:"format"`
	fields map[string]string `toml:"fields"`
}

type Output struct {
	compress   bool   `toml:"compress"`
	console    bool   `toml:"console"`
	file       bool   `toml:"file"`
	filePath   string `toml:"file_path"`
	fileName   string `toml:"file_name"`
	maxAge     int    `toml:"max_age"`
	maxBackups int    `toml:"max_backups"`
	maxSize    int    `toml:"max_size"`
}

// 默认配置
func (conf *Config) DefaultWithModuleName(moduleName string) {
	conf.level = "trace"

	conf.context.format = "console"
	conf.context.fields = map[string]string{}

	conf.output.compress = false
	conf.output.console = true
	conf.output.file = true
	conf.output.filePath = DefaultOutputFolder + moduleName
	conf.output.fileName = DefaultOutputFile
	conf.output.maxAge = 0
	conf.output.maxBackups = 0
	conf.output.maxSize = 10
}

// 修复一些配置文件中可能存在的错误
func (conf *Config) Fix(moduleName string) {
	if conf.output.file {
		if conf.output.filePath == "" {
			conf.output.filePath = DefaultOutputFolder + moduleName
		}
		if conf.output.fileName == "" {
			conf.output.fileName = "fixed_log.txt"
		}
		if conf.output.maxSize == 0 {
			conf.output.maxSize = 10
		}
		if conf.output.maxAge < 0 || conf.output.maxBackups < 0 {
			conf.output.maxAge = 0
			conf.output.maxBackups = 0
		}
	} else {
		conf.output.console = true
	}
	if conf.context.format != "json" && conf.context.format != "console" {
		conf.context.format = "console"
	}
}
