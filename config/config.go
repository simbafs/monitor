package config

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Val[T int | float64] struct {
	Val T
	Def T
	Des string
}

type Config struct {
	int     map[string]Val[int]
	float64 map[string]Val[float64]
}

func New() *Config {
	return &Config{
		int:     make(map[string]Val[int]),
		float64: make(map[string]Val[float64]),
	}
}

// Int defines a new integer configuration value with a default value.
func (c *Config) Int(name string, des string, def int) *Config {
	c.int[name] = Val[int]{
		Val: def,
		Def: def,
		Des: des,
	}
	return c
}

// Float64 defines a new float64 configuration value with a default value.
func (c *Config) Float64(name string, des string, def float64) *Config {
	c.float64[name] = Val[float64]{
		Val: def,
		Def: def,
		Des: des,
	}

	return c
}

// SetInt sets the value of an integer configuration value.
func (c *Config) SetInt(name string, val int) {
	v, ok := c.int[name]
	if !ok {
		return
	}

	v.Val = val

	c.int[name] = v
}

// SetFloat64 sets the value of a float64 configuration value.
func (c *Config) SetFloat64(name string, val float64) {
	v, ok := c.float64[name]
	if !ok {
		return
	}

	v.Val = val

	c.float64[name] = v
}

// ResetInt resets the value of an integer configuration value to its default value.
func (c *Config) ResetInt(name string) {
	v, ok := c.int[name]
	if !ok {
		return
	}

	v.Val = v.Def

	c.int[name] = v
}

// ResetFloat resets the value of a float64 configuration value to its default value.
func (c *Config) ResetFloat(name string) {
	v, ok := c.float64[name]
	if !ok {
		return
	}

	v.Val = v.Def

	c.float64[name] = v
}

// GetInt returns the value of an integer configuration value.
func (c *Config) GetInt(name string) int {
	return c.int[name].Val
}

// GetFloat64 returns the value of a float64 configuration value.
func (c *Config) GetFloat64(name string) float64 {
	return c.float64[name].Val
}

// Cmd handle /set command, if return true, the command was handled.
func (c *Config) Cmd(update tgbotapi.Update) bool {
	msg := update.Message.Text
	fmt.Println(msg)

	return true
}
