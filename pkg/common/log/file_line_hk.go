/*
** description("get the name and line number of the calling file hook").
** copyright('tuoyun,www.tuoyun.net').
** author("fg,Gordon@tuoyun.net").
** time(2021/3/16 11:26).
 */
package log

import (
	"Open_IM/pkg/utils"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

type fileHook struct{}

func newFileHook() *fileHook {
	return &fileHook{}
}

func (f *fileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

//func (f *fileHook) Fire(entry *logrus.Entry) error {
//	entry.Data["FilePath"] = findCaller(6)
//	utils.GetSelfFuncName()
//	return nil
//}

func (f *fileHook) Fire(entry *logrus.Entry) error {
	var s string
	_, file, line, _ := runtime.Caller(8)
	i := strings.SplitAfter(file, "/")
	if len(i) > 3 {
		s = i[len(i)-3] + i[len(i)-2] + i[len(i)-1] + ":" + utils.IntToString(line)
	}
	entry.Data["FilePath"] = s
	return nil
}

//func findCaller(skip int) string {
//	file := ""
//	line := 0
//	for i := 0; i < 10; i++ {
//		file, line = getCaller(skip + i)
//		if !strings.HasPrefix(file, "log") {
//			break
//		}
//	}
//	return fmt.Sprintf("%s:%d", file, line)
//}
//
//func getCaller(skip int) (string, int) {
//	_, file, line, ok := runtime.Caller(skip)
//
//	if !ok {
//		return "", 0
//	}
//	fmt.Println("skip:", skip, "file:", file, "line", line)
//	n := 0
//	for i := len(file) - 1; i > 0; i-- {
//		if file[i] == '/' {
//			n++
//			if n >= 2 {
//				file = file[i+1:]
//				break
//			}
//		}
//	}
//	return file, line
//}
