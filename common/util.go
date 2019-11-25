package common

import (
	"bytes"
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

const (
	logFileName = "stream.log"
	etc_dir = "/etc/"
	data_dir = "/data/"
	log_dir = "/log/"
)

var (
	Log *logrus.Logger
	Config *XStreamConf
	IsTesting bool
	logFile *os.File
	mockTicker *MockTicker
	mockTimer *MockTimer
	mockNow int64
)

type logRedirect struct {

}

func (l *logRedirect) Errorf(f string, v ...interface{}) {
	Log.Error(fmt.Sprintf(f, v...))
}

func (l *logRedirect) Infof(f string, v ...interface{}) {
	Log.Info(fmt.Sprintf(f, v...))
}

func (l *logRedirect) Warningf(f string, v ...interface{}) {
	Log.Warning(fmt.Sprintf(f, v...))
}

func (l *logRedirect) Debugf(f string, v ...interface{}) {
	Log.Debug(fmt.Sprintf(f, v...))
}

func LoadConf(confName string) []byte {
	confDir, err := GetConfLoc()
	if err != nil {
		Log.Fatal(err)
	}

	file := confDir + confName
	b, err := ioutil.ReadFile(file)
	if err != nil {
		Log.Fatal(err)
	}
	return b
}

type XStreamConf struct {
	Debug bool `yaml:"debug"`
	Port int `yaml:"port"`
}

var StreamConf = "kuiper.yaml"

func init(){
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	b := LoadConf(StreamConf)
	var cfg map[string]XStreamConf
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		Log.Fatal(err)
	}

	if c, ok := cfg["basic"]; !ok{
		Log.Fatal("no basic config in kuiper.yaml")
	}else{
		Config = &c
	}

	if !Config.Debug {
		logDir, err := GetLoc(log_dir)
		if err != nil {
			Log.Fatal(err)
		}
		file := logDir + logFileName
		logFile, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			Log.Out = logFile
		} else {
			Log.Infof("Failed to log to file, using default stderr")
		}
	}else{
		Log.SetLevel(logrus.DebugLevel)
	}
}

type KeyValue interface {
	Open() error
	Close() error
	Set(key string, value interface{}) error
	Get(key string) (interface{}, bool)
	Delete(key string) error
	Keys() (keys []string, err error)
}

type SimpleKVStore struct {
	path string
	c *cache.Cache;
}


var stores = make(map[string]*SimpleKVStore)

func GetSimpleKVStore(path string) *SimpleKVStore {
	if s, ok := stores[path]; ok {
		return s
	} else {
		c := cache.New(cache.NoExpiration, 0)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, os.ModePerm)
		}
		sStore := &SimpleKVStore{path: path + "/stores.data", c: c}
		stores[path] = sStore
		return sStore
	}
}

func (m *SimpleKVStore) Open() error  {
	if _, err := os.Stat(m.path); os.IsNotExist(err) {
		return nil
	}
	if e := m.c.LoadFile(m.path); e != nil {
		return e
	}
	return nil
}

func (m *SimpleKVStore) Close() error  {
	e := m.saveToFile()
	m.c.Flush() //Delete all of the values from memory.
	return e
}

func (m *SimpleKVStore) saveToFile() error {
	if e := m.c.SaveFile(m.path); e != nil {
		return e
	}
	return nil
}

func (m *SimpleKVStore) Set(key string, value interface{}) error  {
	if m.c == nil {
		return fmt.Errorf("Cache %s has not been initialized yet.", m.path)
	}
	m.c.Set(key, value, cache.NoExpiration)
	return m.saveToFile()
}

func (m *SimpleKVStore) Get(key string) (interface{}, bool)  {
	return m.c.Get(key)
}

func (m *SimpleKVStore) Delete(key string) error {
	m.c.Delete(key)
	return m.saveToFile()
}

func (m *SimpleKVStore) Keys() (keys []string, err error) {
	if m.c == nil {
		return nil, fmt.Errorf("Cache %s has not been initialized yet.", m.path)
	}
	its := m.c.Items()
	keys = make([]string, 0, len(its))
	for k := range its {
		keys = append(keys, k)
	}
	return keys, nil
}

func PrintMap(m map[string]string, buff *bytes.Buffer) {

	for k, v := range m {
		buff.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
}

func CloseLogger(){
	if logFile != nil {
		logFile.Close()
	}
}

func GetConfLoc()(string, error){
	return GetLoc(etc_dir)
}

func GetDataLoc() (string, error) {
	return GetLoc(data_dir)
}

func GetLoc(subdir string)(string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	confDir := dir + subdir
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		lastdir := dir
		for len(dir) > 0 {
			dir = filepath.Dir(dir)
			if lastdir == dir {
				break
			}
			confDir = dir + subdir
			if _, err := os.Stat(confDir); os.IsNotExist(err) {
				lastdir = dir
				continue
			} else {
				//Log.Printf("Trying to load file from %s", confDir)
				return confDir, nil
			}
		}
	} else {
		//Log.Printf("Trying to load file from %s", confDir)
		return confDir, nil
	}

	return "", fmt.Errorf("conf dir not found")
}

func GetAndCreateDataLoc(dir string) (string, error) {
	dataDir, err := GetDataLoc()
	if err != nil {
		return "", err
	}
	d := path.Join(path.Dir(dataDir), dir)
	if _, err := os.Stat(d); os.IsNotExist(err) {
		err = os.MkdirAll(d, 0755)
		if err != nil {
			return "", err
		}
	}
	return d, nil
}

//Time related. For Mock
func GetTicker(duration int) Ticker {
	if IsTesting{
		if mockTicker == nil{
			mockTicker = NewMockTicker(duration)
		}else{
			mockTicker.SetDuration(duration)
		}
		return mockTicker
	}else{
		return NewDefaultTicker(duration)
	}
}

func GetTimer(duration int) Timer {
	if IsTesting{
		if mockTimer == nil{
			mockTimer = NewMockTimer(duration)
		}else{
			mockTimer.SetDuration(duration)
		}
		return mockTimer
	}else{
		return NewDefaultTimer(duration)
	}
}

func ProcessPath(p string) (string, error) {
	if abs, err := filepath.Abs(p); err != nil {
		return "", nil
	} else {
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			return "", err;
		}
		return abs, nil
	}
}

/****** For Test Only ********/
func GetMockTicker() *MockTicker{
	return mockTicker
}

func ResetMockTicker(){
	if mockTicker != nil{
		mockTicker.lastTick = 0
	}
}

func GetMockTimer() *MockTimer{
	return mockTimer
}

func SetMockNow(now int64){
	mockNow = now
}

func GetMockNow() int64{
	return mockNow
}

/*********** Type Cast Utilities *****/
//TODO datetime type
func ToString(input interface{}) string{
	return fmt.Sprintf("%v", input)
}
func ToInt(input interface{}) (int, error){
	switch t := input.(type) {
	case float64:
		return int(t), nil
	case int64:
		return int(t), nil
	case int:
		return t, nil
	default:
		return 0, fmt.Errorf("unsupported type %T of %[1]v", input)
	}
}
