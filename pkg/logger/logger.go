package logger

import "log"

func Info(msg string, kv ...interface{}) { log.Println(append([]interface{}{"INFO:", msg}, kv...)...) }
func Error(msg string, kv ...interface{}) { log.Println(append([]interface{}{"ERROR:", msg}, kv...)...) }
