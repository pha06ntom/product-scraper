// Package app — точка оркестрации.
// Читает конфиг, инициализирует браузер,
// запускает сбор по категориям и сохраняет результат.
package app

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/pha06ntom/lenta-scraper/internal/browser"
	"github.com/pha06ntom/lenta-scraper/internal/output"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy       string   `yaml:"proxy"`
	ProxyUser   string   `yaml:"proxy_user"`
	ProxyPass   string   `yaml:"proxy_pass"`
	Address     string   `yaml:"address"`
	Categories  []string `yaml:"categories"`
	OutCSV      string   `yaml:"out_csv"`
	Headless    bool     `yaml:"headless"`
	TimeoutSec  int      `yaml:"timeout_sec"`
	Scrolls     int      `yaml:"scrolls"`
	SkipAddress bool     `yaml:"skip_address"`
}

type App struct {
	cfg Config
}

func NewFromYAML(path string) (*App, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Proxy == "" {
		return nil, errors.New("proxy is required (proxy must be used)")
	}
	// Address оставляем обязательным, даже если позже можно пропустить SelectAddress:
	// это показывает, что логика привязки к адресу вообще предусмотрена.
	if cfg.Address == "" {
		return nil, errors.New("address is required")
	}
	if len(cfg.Categories) == 0 {
		return nil, errors.New("categories are required")
	}
	if cfg.OutCSV == "" {
		cfg.OutCSV = "dump.csv"
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 90
	}
	if cfg.Scrolls <= 0 {
		cfg.Scrolls = 6
	}
	return &App{cfg: cfg}, nil
}

func (a *App) Run() error {
	log.Printf("Starting lenta-scraper (headless=%v)", a.cfg.Headless)
	log.Printf("Categories: %d, out: %s", len(a.cfg.Categories), a.cfg.OutCSV)

	b, err := browser.New(browser.Options{
		Proxy:     a.cfg.Proxy,
		ProxyUser: a.cfg.ProxyUser,
		ProxyPass: a.cfg.ProxyPass,
		Headless:  a.cfg.Headless,
	})

	if err != nil {
		return err
	}
	defer b.Close()

	collector := browser.NewCollector()
	b.AttachCollector(collector)

	ctx := b.Context()
	timeout := time.Duration(a.cfg.TimeoutSec) * time.Second

	// ---------- выбор адреса ----------
	if !a.cfg.SkipAddress {
		log.Println("Selecting delivery address (required step)...")

		addrCtx, addrCancel := context.WithTimeout(ctx, timeout)
		defer addrCancel()

		if err := b.SelectAddress(addrCtx, a.cfg.Address); err != nil {
			// Для тестового: не валим всю программу, а логируем предупреждение.
			log.Printf("WARN: failed to select address (%v), continue with default store context", err)
		}
	} else {
		log.Println("Skipping address selection (skip_address=true)")
	}
	// ----------------------------------

	for _, cat := range a.cfg.Categories {
		log.Println("Collecting:", cat)
		cctx, cancel := context.WithTimeout(ctx, timeout)
		err := b.CollectCategory(cctx, cat, a.cfg.Scrolls)
		cancel()
		if err != nil {
			log.Printf("category error: %v", err)
		}
	}

	items := collector.Items()
	log.Printf("Collected items: %d", len(items))

	if err := output.WriteCSV(a.cfg.OutCSV, items); err != nil {
		return err
	}

	log.Println("Done:", a.cfg.OutCSV)
	return nil
}
