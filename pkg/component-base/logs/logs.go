package logs

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ILogger interface {
	Trace() *zerolog.Event
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Panic() *zerolog.Event
}

type Logger struct {
	logger zerolog.Logger
}

var logs Logger

func Init(moduleName string) {
	conf := Config{}
	conf.DefaultWithModuleName(moduleName)
	initLogs(conf)
}

func initLogs(conf Config) {
	// 为日志文件名添加日期
	currentTime := time.Now().Format(time.DateOnly)
	conf.output.fileName = currentTime + "_" + conf.output.fileName

	// 设置日志轮转
	var writers []io.Writer
	logFile := filepath.Join(conf.output.filePath, conf.output.fileName)
	if conf.output.file {
		hook := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    conf.output.maxSize,
			MaxBackups: conf.output.maxBackups,
			MaxAge:     conf.output.maxAge,
			Compress:   conf.output.compress,
			LocalTime:  true,
		}

		writers = append(writers, bufio.NewWriter(hook))
	}

	// 确保日志文件夹存在
	logDir := checkSavePath(conf.output.filePath)
	// 检查并打开日志文件
	// file := OpenLogFile(logDir, conf.output.fileName)
	logFilePath := filepath.Join(logDir, conf.output.fileName)
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	if conf.output.file {
		writers = append(writers, file) // 添加文件输出
	}
	if conf.output.console {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout,
			TimeFormat: time.DateTime,
		}) // 也可以只输出Stderr
	}

	// 创建一个日志
	multi := zerolog.MultiLevelWriter(writers...)
	logs.logger = zerolog.New(multi).With().Timestamp().Stack().Logger()
	switch conf.level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// 添加配置文件中的字段
	context := logs.logger.With()
	for key, value := range conf.context.fields {
		context = context.Str(key, value)
	}
	logs.logger = context.Logger()
}

/*------------------------------------辅助函数-----------------------------------------*/
/*
返回： 模块名 、 模块相对路径 、 错误的行数
*/
func getModuleName(skipLevel int) (string, string, int) {
	_, file, line, ok := runtime.Caller(skipLevel)
	// fmt.Println(file, line)
	var moduleName, modulePath string
	if !ok {
		return "unknownModuleName", "unknownModulePath", 0
	}

	// 获取模块名
	startIndex := strings.Index(file, "pkg/")
	// fmt.Println("startIndex:", startIndex)
	subPath := file[startIndex+len("pkg/"):]
	// fmt.Println("subPath:" + subPath)
	endIndex := strings.Index(subPath, "/")
	if endIndex == -1 {
		moduleName = "unknownModuleName"
	} else {
		moduleName = subPath[:endIndex]
	}
	// fmt.Println(moduleName)

	//获取模块的相对路径
	workPath, err := os.Getwd()
	if err != nil {
		modulePath = "unknownModulePath"
	} else {
		relativePath, err := filepath.Rel(workPath, file)
		if err != nil {
			modulePath = "unknownModulePath"
		} else {
			modulePath = strings.ReplaceAll(relativePath, string(filepath.Separator), "/")
		}
	}
	return moduleName, modulePath, line
}

// 检查日志的文件夹是否存在，不存在创建一个
func checkSavePath(logDir string) string {
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			panic(err)
		}
	}
	return logDir
}

func (l *Logger) addContext(context zerolog.Context, fields ...interface{}) zerolog.Context {
	// 添加调用者的信息
	_, modulePath, line := getModuleName(3)
	modulePathWithLine := modulePath + ":" + strconv.Itoa(line)
	context = context.Str("module", modulePathWithLine)

	// 添加传入的信息
	var stringMessage string
	if len(fields) > 0 {
		for _, field := range fields {
			if _, ok := field.(string); ok {
				if stringMessage != "" {
					stringMessage += " | "
				}
				stringMessage += field.(string)
			} else if _, ok := field.(int); ok {
				continue
			} else if _, ok := field.(map[string]interface{}); ok {
				// context = context.Fields(field)
				context = context.Fields(flattenMap(field.(map[string]interface{}), "", map[string]interface{}{}))
			} else {
				context = context.Fields(parseStructFields(field))
			}
		}
	}
	if stringMessage != "" {
		context = context.Str("msg", stringMessage)
	}
	return context
}

func flattenMap(data map[string]interface{}, parentKey string, result map[string]interface{}) map[string]interface{} {
	for k, v := range data {
		fullKey := k
		if parentKey != "" {
			fullKey = parentKey + "." + k
		}
		switch v := v.(type) {
		case map[string]interface{}:
			flattenMap(v, fullKey, result)
		case map[string]string:
			for subKey, subValue := range v {
				result[fullKey+"."+subKey] = subValue
			}
		case map[string]int:
			for subKey, subValue := range v {
				result[fullKey+"."+subKey] = subValue
			}
		default:
			result[fullKey] = v
		}
	}
	return result
}

func Trace(fields ...interface{}) {
	context := logs.logger.With()
	context = logs.addContext(context, fields...)
	newLogger := context.Logger()
	logger := &newLogger
	logger.Trace().Msgf("")
}

func Tracef(format string, v ...interface{}) {
	logs.logger.Trace().Msgf(format, v...)
}

func Debug(fields ...interface{}) {
	context := logs.logger.With()
	context = logs.addContext(context, fields...)
	newLogger := context.Logger()
	logger := &newLogger
	logger.Debug().Msgf("")
}

func Debugf(format string, v ...interface{}) {
	logs.logger.Debug().Msgf(format, v...)
}

func Info(fields ...interface{}) {
	context := logs.logger.With()
	context = logs.addContext(context, fields...)
	newLogger := context.Logger()
	logger := &newLogger
	logger.Info().Msgf("")
}

func Infof(format string, v ...interface{}) {
	logs.logger.Info().Msgf(format, v...)
}

func Warn(fields ...interface{}) {
	context := logs.logger.With()
	context = logs.addContext(context, fields...)
	newLogger := context.Logger()
	logger := &newLogger
	logger.Warn().Msgf("")
}

func Warnf(format string, v ...interface{}) {
	logs.logger.Warn().Msgf(format, v...)
}

func Error(fields ...interface{}) {
	context := logs.logger.With()
	context = logs.addContext(context, fields...)
	newLogger := context.Logger()
	logger := &newLogger
	logger.Error().Msgf("")
}

func Errorf(format string, v ...interface{}) {
	logs.logger.Error().Msgf(format, v...)
}

func Fatal(fields ...interface{}) {
	context := logs.logger.With()
	context = logs.addContext(context, fields...)
	newLogger := context.Logger()
	logger := &newLogger
	logger.Fatal().Msgf("")
}

func Fatalf(format string, v ...interface{}) {
	logs.logger.Fatal().Msgf(format, v...)
}

func Panic(fields ...interface{}) {
	context := logs.logger.With()
	context = logs.addContext(context, fields...)
	newLogger := context.Logger()
	logger := &newLogger
	logger.Panic().Msgf("")
}

func Panicf(format string, v ...interface{}) {
	logs.logger.Panic().Msgf(format, v...)
}

func parseStructFields(v interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	// 使用反射获取结构体类型和字段
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)
	// 确保 v 是一个结构体类型
	if val.Kind() == reflect.Struct {
		// 遍历结构体字段
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldName := typ.Field(i).Name

			// 动态构建字段名，使用结构体类型名 + 字段名
			uniqueFieldName := fmt.Sprintf("%s.%s", typ.Name(), fieldName)

			// 将字段名和值添加到字段集合中
			fields[uniqueFieldName] = field.Interface()
		}
	}
	return fields
}
