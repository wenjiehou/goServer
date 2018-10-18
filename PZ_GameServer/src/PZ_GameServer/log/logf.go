package log

// 日志类

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// levels
const (
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
)

const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printErrorLevel   = "[error  ] "
	printFatalLevel   = "[fatal  ] "
)

var (
	ToFile    = false  //是否输出到文件
	Level     = 0      //输出等级
	DebugPath = "log/" //默认输出路径
	logger    *Logger
)

type Logger struct {
	level      int
	baseLogger *log.Logger
	baseFile   *os.File
}

func InitConfig(lev string, path string, tofile bool) {
	ToFile = tofile
	DebugPath = path

	var level int
	switch strings.ToLower(lev) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		level = debugLevel
		return
	}

	Level = level

	if ToFile {
		logger = NewFile(Level, DebugPath, log.LstdFlags)
	}
}

////新建日志文件
func NewFile(level int, pathname string, flag int) *Logger {

	if ToFile == false {
		return nil
	}

	// logger
	var baseLogger *log.Logger
	var baseFile *os.File
	if pathname != "" {
		now := time.Now()

		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())

		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			fmt.Println("err create log file : " + path.Join(pathname, filename))
			return nil
		}

		baseLogger = log.New(file, "", flag)
		baseFile = file

		//fmt.Println("log file in : " + path.Join(pathname, filename) + " created!")

	} else {
		baseLogger = log.New(os.Stdout, "", flag)
	}

	// new
	lg := new(Logger)
	lg.level = level
	lg.baseLogger = baseLogger
	lg.baseFile = baseFile

	return lg

}

//func NewFile(strLevel string, pathname string, flag int) (*Logger, error) {
//	// level
//	var level int
//	switch strings.ToLower(strLevel) {
//	case "debug":
//		level = debugLevel
//	case "release":
//		level = releaseLevel
//	case "error":
//		level = errorLevel
//	case "fatal":
//		level = fatalLevel
//	default:
//		return nil, errors.New("unknown level: " + strLevel)
//	}

//	// logger
//	var baseLogger *log.Logger
//	var baseFile *os.File
//	if pathname != "" {
//		now := time.Now()

//		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
//			now.Year(),
//			now.Month(),
//			now.Day(),
//			now.Hour(),
//			now.Minute(),
//			now.Second())

//		file, err := os.Create(path.Join(pathname, filename))
//		if err != nil {
//			return nil, err
//		}

//		baseLogger = log.New(file, "", flag)
//		baseFile = file
//	} else {
//		baseLogger = log.New(os.Stdout, "", flag)
//	}

//	// new
//	logger := new(Logger)
//	logger.level = level
//	logger.baseLogger = baseLogger
//	logger.baseFile = baseFile

//	return logger, nil
//}

//// It's dangerous to call the method on logging
//func (logger *Logger) Close() {
//	if logger.baseFile != nil {
//		logger.baseFile.Close()
//	}

//	logger.baseLogger = nil
//	logger.baseFile = nil
//}

func doPrintf(level int, printLevel string, format string, a ...interface{}) {

	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}

	format = printLevel + format
	logger.baseLogger.Output(3, fmt.Sprintf(format, a...))

	if level == fatalLevel {
		os.Exit(1)
	}
}

func Debug(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
	if ToFile && logger != nil {
		doPrintf(debugLevel, printDebugLevel, format, a...)
	}
}

func Release(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
	if ToFile && logger != nil {
		doPrintf(releaseLevel, printReleaseLevel, format, a...)
	}
}

func Error(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
	if ToFile && logger != nil {
		doPrintf(errorLevel, printErrorLevel, format, a...)
	}
}

func Fatal(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf(format, a...))
	if ToFile && logger != nil {
		doPrintf(fatalLevel, printFatalLevel, format, a...)
	}
}

//// It's dangerous to call the method on logging
//func Export(logger *Logger) {
//	if logger != nil {
//		Logf = logger
//	}
//}

//func Debug(format string, a ...interface{}) {
//	Logf.doPrintf(debugLevel, printDebugLevel, format, a...)
//}

//func Release(format string, a ...interface{}) {
//	Logf.doPrintf(releaseLevel, printReleaseLevel, format, a...)
//}

//func Error(format string, a ...interface{}) {
//	Logf.doPrintf(errorLevel, printErrorLevel, format, a...)
//}

//func Fatal(format string, a ...interface{}) {
//	Logf.doPrintf(fatalLevel, printFatalLevel, format, a...)
//}

func Close() {
	if logger != nil {
		logger.baseFile.Close()
	}
}
