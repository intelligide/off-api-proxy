package config

import (
	"errors"
	"github.com/astaxie/beego/config"
	"github.com/pelletier/go-toml"
	"os"
	"strings"
	"sync"
)

func init() {
	config.Register("toml", &TOMLConfig{})
}

// JSONConfig is a toml config parser and implements Config interface.
type TOMLConfig struct {}

func (js *TOMLConfig) Parse(filename string) (config.Configer, error) {
	tomlTree, err := toml.LoadFile(filename)
	if err != nil {
		return nil, err
	}

	x := &TOMLConfigContainer{
		data: tomlTree,
	}

	return x, nil
}

func (js *TOMLConfig) ParseData(data []byte) (config.Configer, error) {
	tomlTree, err := toml.LoadBytes(data)
	if err != nil {
		return nil, err
	}

	x := &TOMLConfigContainer{
		data: tomlTree,
	}

	return x, nil
}

type TOMLConfigContainer struct {
	data *toml.Tree
	sync.RWMutex
}

func (c *TOMLConfigContainer) Set(key, val string) error {
	c.Lock()
	defer c.Unlock()
	c.data.Set(key, val)
	return nil
}

func (c *TOMLConfigContainer) String(key string) string {
	val := c.data.Get(key)
	if val != nil {
		if v, ok := val.(string); ok {
			return v
		}
	}
	return ""
}

func (c *TOMLConfigContainer) DefaultString(key string, defaultVal string) string {
	if v := c.String(key); len(v) > 0 {
		return v
	}
	return defaultVal
}

func (c *TOMLConfigContainer) Strings(key string) []string {
	stringVal := c.String(key)
	if stringVal == "" {
		return nil
	}
	return strings.Split(c.String(key), ";")
}

func (c *TOMLConfigContainer) DefaultStrings(key string, defaultVal []string) []string {
	if v := c.Strings(key); v != nil {
		return v
	}
	return defaultVal
}

func (c *TOMLConfigContainer) Int(key string) (int, error) {
	val := c.data.Get(key)
	if val == nil {
		return 0, errors.New("not exist key:" + key)
	}

	if v, ok := val.(int); ok {
		return v, nil
	}

	return 0, errors.New("not int value")
}

func (c *TOMLConfigContainer) DefaultInt(key string, defaultVal int) int {
	if v, err := c.Int(key); err == nil {
		return v
	}
	return defaultVal
}

func (c *TOMLConfigContainer) Int64(key string) (int64, error) {
	val := c.data.Get(key)
	if val == nil {
		return 0, errors.New("not exist key:" + key)
	}

	if v, ok := val.(int64); ok {
		return v, nil
	}

	return 0, errors.New("not int64 value")
}

func (c *TOMLConfigContainer) DefaultInt64(key string, defaultVal int64) int64 {
	if v, err := c.Int64(key); err == nil {
		return v
	}
	return defaultVal
}

func (c *TOMLConfigContainer) Bool(key string) (bool, error) {
	val := c.data.Get(key)
	if val == nil {
		return false, errors.New("not exist key:" + key)
	}

	if v, ok := val.(bool); ok {
		return v, nil
	}

	return false, errors.New("not bool value")
}

func (c *TOMLConfigContainer) DefaultBool(key string, defaultVal bool) bool {
	if v, err := c.Bool(key); err == nil {
		return v
	}
	return defaultVal
}

func (c *TOMLConfigContainer) Float(key string) (float64, error) {
	val := c.data.Get(key)
	if val == nil {
		return 0, errors.New("not exist key:" + key)
	}

	if v, ok := val.(float64); ok {
		return v, nil
	}

	return 0, errors.New("not float64 value")
}

func (c *TOMLConfigContainer) DefaultFloat(key string, defaultVal float64) float64 {
	if v, err := c.Float(key); err == nil {
		return v
	}
	return defaultVal
}

func (c *TOMLConfigContainer) DIY(key string) (interface{}, error) {
	val := c.data.Get(key)
	if val == nil {
		return 0, errors.New("not exist key:" + key)
	}
	return val, nil
}

func (c *TOMLConfigContainer) GetSection(section string) (map[string]string, error) {
	val := c.data.Get(section)
	if val != nil {
		if v, ok := val.(map[string]string); ok {
			return v, nil
		}
	}
	return nil, errors.New("not float64 value")
}

func (c *TOMLConfigContainer) SaveConfigFile(filename string) error {
	// Write configuration file by filename.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = c.data.WriteTo(f)
	return err
}
