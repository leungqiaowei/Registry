package logs

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	/**
	检查了init以及initLogs ， init里面会调用initLogs
	检查是否在指定的地方建立日志文件夹和日志文件，以及日志文件能够正常写入文件中
	检查当前的log是否有字段time、module
	*/
	moduleName := "testModule"
	Init(moduleName)

	// 验证日志文件是否创建
	currentTime := time.Now().Format("2006-01-02")
	expectedLogFile := filepath.Join("/tmp/logs/testModule", currentTime+"_log.txt")
	_, err := os.Stat(expectedLogFile)
	assert.NoError(t, err, "日志文件应存在: %s", expectedLogFile)

	// 验证日志内容
	file, err := os.Open(expectedLogFile)
	assert.NoError(t, err, "无法打开日志文件: %s", expectedLogFile)

	Trace()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(file)
	assert.NoError(t, err, "无法读取日志文件内容: %s", expectedLogFile)

	logContent := buf.String()
	assert.Contains(t, logContent, "module", "日志内容不包含module字段")
	assert.Contains(t, logContent, "time", "日志内容不包含time字段")
}

func TestGetModuleName(t *testing.T) {
	/**
	根据这个函数调用getModuleName来检查获取的模块名、相对路径以及调用的行数是否正确
	*/
	// 下面两行中间不要插入其他的东西，才能使得调用行数的判断正确
	moduleName, modulePath, line := getModuleName(1)
	_, _, expectLine, ok := runtime.Caller(0)
	if ok {
		if expectLine-1 != line {
			t.Errorf("expect line %d, got %d", expectLine-1, line)
		}
		if moduleName != "component-base" {
			t.Errorf("expect module name component-base, got %s", moduleName)
		}
		if modulePath != "logs_test.go" {
			t.Errorf("expect module path logs_test.go , got %s", modulePath)
		}

	}
}

type SampleStruct struct {
	Name  string
	Age   int
	Email string
}

func TestCheckSavePath(t *testing.T) {
	/**
	检查这个函数是否能够成功的创建一个不存在的文件夹
	*/
	// 创建一个临时目录作为测试环境
	tempDir := t.TempDir()

	// 测试目录不存在时，函数创建它
	testPath := filepath.Join(tempDir, "new_log_dir")
	checkedPath := checkSavePath(testPath)

	// 检查目录是否存在
	_, err := os.Stat(checkedPath)
	assert.NoError(t, err, "Expected directory to be created")
}

func TestFlattenMap(t *testing.T) {
	a := map[string]interface{}{
		"name":   "cxl",
		"age":    "18",
		"school": "tju",
		"likes": map[string]string{
			"music": "1",
			"draw":  "2",
			"dance": "0",
		},
		"dengji": map[string]interface{}{
			"haha": map[string]int{
				"lala": 11,
				"baba": 34,
			},
			"nana": 2,
		},
	}

	result := flattenMap(a, "", make(map[string]interface{}))

	expected := map[string]interface{}{
		"name":             "cxl",
		"age":              "18",
		"school":           "tju",
		"likes.music":      "1",
		"likes.draw":       "2",
		"likes.dance":      "0",
		"dengji.haha.lala": 11,
		"dengji.haha.baba": 34,
		"dengji.nana":      2,
	}

	// 检查结果是否符合预期
	for key, expectedValue := range expected {
		if value, exists := result[key]; !exists || value != expectedValue {
			t.Errorf("Field %q: expected %v, got %v", key, expectedValue, value)
		}
	}

	// 检查是否有额外的字段
	for key := range result {
		if _, exists := expected[key]; !exists {
			t.Errorf("Unexpected field %q in result", key)
		}
	}
}

type testCase struct {
	level string
}

func createTestCase(t *testing.T, c *testCase) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()

	a := map[string]interface{}{
		"name":   "cxl",
		"age":    "18",
		"school": "tju",
		"likes": map[string]string{"music": "1",
			"draw":  "2",
			"dance": "0",
		},
		"dengji": map[string]interface{}{
			"haha": map[string]int{
				"lala": 11,
				"baba": 34,
			},
			"nana": 2,
		},
	}
	fields := []interface{}{
		a,
		"This is message",
		SampleStruct{Name: "haha",
			Age:   12,
			Email: "293@email.com"}}
	switch c.level {
	case "debug":
		Debug(fields...)
	case "info":
		Info(fields...)
	case "warn":
		Warn(fields...)
	case "error":
		Error(fields...)
	case "trace":
		Trace(fields...)
	}
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	//logOutput := buf.String()
	//fmt.Println(logOutput)

	expectedFields := map[string]interface{}{
		"level":              c.level,
		"module":             "logs_test.go:xx",
		"SampleStruct.Name":  "haha",
		"SampleStruct.Age":   float64(12),
		"SampleStruct.Email": "293@email.com",
		"msg":                "This is message",
		"name":               "cxl",
		"age":                "18",
		"school":             "tju",
		"likes.music":        "1",
		"likes.draw":         "2",
		"likes.dance":        "0",
		"dengji.haha.lala":   float64(11),
		"dengji.haha.baba":   float64(34),
		"dengji.nana":        float64(2),
	}

	// 验证日志内容
	for key, expectedValue := range expectedFields {
		// 验证时间字段存在
		if _, exists := logData["time"]; !exists {
			t.Error("Expected time field missing in log output")
		}
		// 检查 module 是否存在
		if key == "module" {
			if _, exists := logData[key]; !exists {
				t.Errorf("Expected key %q missing in log output", key)
			}
			continue
		}
		// 检验其他字段是否存在以及值是否正确
		if value, exists := logData[key]; !exists || value != expectedValue {
			t.Errorf("Field %q: expected %v, got %v", key, expectedValue, value)
		}
	}

}

func TestTracef(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()
	Tracef("Test %s format %v ", "Trace", 1)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if value, exists := logData["message"]; !exists || value != "Test Trace format 1 " {
		t.Errorf("没有成功输出字符串")
	}
}

func TestTrace(t *testing.T) {
	createTestCase(t, &testCase{"trace"})
}

func TestDebugf(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()
	Debugf("Test %s format %v ", "Debug", 1)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if value, exists := logData["message"]; !exists || value != "Test Debug format 1 " {
		t.Errorf("没有成功输出字符串")
	}
}

func TestDebug(t *testing.T) {
	createTestCase(t, &testCase{"debug"})
}

func TestInfof(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()
	Infof("Test %s format %v ", "Info", 1)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if value, exists := logData["message"]; !exists || value != "Test Info format 1 " {
		t.Errorf("没有成功输出字符串")
	}
}

func TestInfo(t *testing.T) {
	createTestCase(t, &testCase{"info"})
}

func TestWarnf(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()
	Warnf("Test %s format %v ", "Warn", 1)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if value, exists := logData["message"]; !exists || value != "Test Warn format 1 " {
		t.Errorf("没有成功输出字符串")
	}
}

func TestWarn(t *testing.T) {
	createTestCase(t, &testCase{"warn"})
}

func TestErrorf(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()
	Errorf("Test %s format %v ", "Error", 1)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if value, exists := logData["message"]; !exists || value != "Test Error format 1 " {
		t.Errorf("没有成功输出字符串")
	}
}

func TestError(t *testing.T) {
	createTestCase(t, &testCase{"error"})
}

// zerolog.Fatal 在调用Msg或者Msgf的时候会执行os.Exit(1)立刻终止函数—— 使用zerolog.Hook 机制模拟Fatal的行为
type ExitHook struct{}

func (h ExitHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level == zerolog.FatalLevel {
		panic("simulate fatal")
	}
}

func TestFatal(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger().Hook(ExitHook{})

	defer func() {
		if r := recover(); r == nil {
			t.Log("Recovered from fatal:", r)
		}
	}()

	a := map[string]interface{}{
		"name":   "cxl",
		"age":    "18",
		"school": "tju",
		"likes": map[string]string{"music": "1",
			"draw":  "2",
			"dance": "0",
		},
		"dengji": map[string]interface{}{
			"haha": map[string]int{
				"lala": 11,
				"baba": 34,
			},
			"nana": 2,
		},
	}
	fields := []interface{}{
		a,
		"This is message",
		SampleStruct{Name: "haha",
			Age:   12,
			Email: "293@email.com"}}
	Fatal(fields...)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	//logOutput := buf.String()
	//fmt.Println(logOutput)

	expectedFields := map[string]interface{}{
		"level":              "fatal",
		"module":             "logs_test.go:xx",
		"SampleStruct.Name":  "haha",
		"SampleStruct.Age":   float64(12),
		"SampleStruct.Email": "293@email.com",
		"msg":                "This is message",
		"name":               "cxl",
		"age":                "18",
		"school":             "tju",
		"likes.music":        "1",
		"likes.draw":         "2",
		"likes.dance":        "0",
		"dengji.haha.lala":   float64(11),
		"dengji.haha.baba":   float64(34),
		"dengji.nana":        float64(2),
	}

	// 验证日志内容
	for key, expectedValue := range expectedFields {
		// 验证时间字段存在
		if _, exists := logData["time"]; !exists {
			t.Error("Expected time field missing in log output")
		}
		// 检查 module 是否存在
		if key == "module" {
			if _, exists := logData[key]; !exists {
				t.Errorf("Expected key %q missing in log output", key)
			}
			continue
		}
		// 检验其他字段是否存在以及值是否正确
		if value, exists := logData[key]; !exists || value != expectedValue {
			t.Errorf("Field %q: expected %v, got %v", key, expectedValue, value)
		}
	}
}

func TestFatalf(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger().Hook(ExitHook{})
	defer func() {
		if r := recover(); r == nil {
			t.Log("Recovered from fatal:", r)
		}
	}()
	Fatalf("Test %s format %v ", "Fatal", 1)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if value, exists := logData["message"]; !exists || value != "Test Fatal format 1 " {
		t.Errorf("没有成功输出字符串")
	}
}

func TestPanic(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()

	defer func() {
		if r := recover(); r == nil {
			t.Log("Recovered from panic:", r)
		}
	}()

	a := map[string]interface{}{
		"name":   "cxl",
		"age":    "18",
		"school": "tju",
		"likes": map[string]string{"music": "1",
			"draw":  "2",
			"dance": "0",
		},
		"dengji": map[string]interface{}{
			"haha": map[string]int{
				"lala": 11,
				"baba": 34,
			},
			"nana": 2,
		},
	}
	fields := []interface{}{
		a,
		"This is message",
		SampleStruct{Name: "haha",
			Age:   12,
			Email: "293@email.com"}}
	Panic(fields...)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}
	//logOutput := buf.String()
	//fmt.Println(logOutput)

	expectedFields := map[string]interface{}{
		"level":              "panic",
		"module":             "logs_test.go:xx",
		"SampleStruct.Name":  "haha",
		"SampleStruct.Age":   float64(12),
		"SampleStruct.Email": "293@email.com",
		"msg":                "This is message",
		"name":               "cxl",
		"age":                "18",
		"school":             "tju",
		"likes.music":        "1",
		"likes.draw":         "2",
		"likes.dance":        "0",
		"dengji.haha.lala":   float64(11),
		"dengji.haha.baba":   float64(34),
		"dengji.nana":        float64(2),
	}

	// 验证日志内容
	for key, expectedValue := range expectedFields {
		// 验证时间字段存在
		if _, exists := logData["time"]; !exists {
			t.Error("Expected time field missing in log output")
		}
		// 检查 module 是否存在
		if key == "module" {
			if _, exists := logData[key]; !exists {
				t.Errorf("Expected key %q missing in log output", key)
			}
			continue
		}
		// 检验其他字段是否存在以及值是否正确
		if value, exists := logData[key]; !exists || value != expectedValue {
			t.Errorf("Field %q: expected %v, got %v", key, expectedValue, value)
		}
	}
}

func TestPanicf(t *testing.T) {
	var buf bytes.Buffer
	logs.logger = zerolog.New(&buf).With().Timestamp().Logger()
	defer func() {
		if r := recover(); r == nil {
			t.Log("Recovered from panic:", r)
		}
	}()
	Panicf("Test %s format %v ", "Panic", 1)
	var logData map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logData)
	if err != nil {
		t.Fatalf("Failed to parse log output: %v", err)
	}

	if value, exists := logData["message"]; !exists || value != "Test Panic format 1 " {
		t.Errorf("没有成功输出字符串")
	}
}

func TestParseStructFields(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected map[string]interface{}
	}{
		{
			name: "ValidStruct",
			input: SampleStruct{
				Name:  "Alice",
				Age:   30,
				Email: "alice@example.com",
			},
			expected: map[string]interface{}{
				"SampleStruct.Name":  "Alice",
				"SampleStruct.Age":   30,
				"SampleStruct.Email": "alice@example.com",
			},
		},
		{
			name:     "NonStructInput",
			input:    123,
			expected: map[string]interface{}{}, // 非结构体输入应返回空 map
		},
		{
			name:     "NilInput",
			input:    nil,
			expected: map[string]interface{}{}, // nil 输入应返回空 map
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseStructFields(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseStructFields(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}
