// Package browser отвечает за управление headless Chrome,
// настройку прокси, выбор адреса доставки и переход по категориям.
//
// Вся работа с chromedp изолирована здесь,
// чтобы бизнес-логика (extract/output) не зависела от конкретной реализации браузера.
package browser

import (
	"context"
	"errors"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"
)

// Options описывает параметры запуска браузера.
type Options struct {
	Proxy     string // адрес прокси, например: http://45.11.21.182:3000
	ProxyUser string // логин для прокси (Proxy6)
	ProxyPass string // пароль для прокси (Proxy6)
	Headless  bool   // запускать ли браузер в headless-режиме
}

type Browser struct {
	allocCtx context.Context
	ctx      context.Context
	cancel   context.CancelFunc

	collector *Collector
}

func New(opts Options) (*Browser, error) {
	// Прокси обязателен по требованиям задания.
	if opts.Proxy == "" {
		return nil, errors.New("proxy is required")
	}

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(opts.Proxy),
		chromedp.Flag("headless", opts.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true), // <-- исправлено
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("lang", "ru-RU"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	// Если заданы логин/пароль прокси — включаем перехват auth-запросов.
	if opts.ProxyUser != "" {
		// Включаем CDP-домен Fetch, чтобы ловить события AuthRequired.
		if err := chromedp.Run(ctx, fetch.Enable()); err != nil {
			cancel()
			allocCancel()
			return nil, err
		}

		// Слушаем события авторизации и отвечаем кредами прокси.
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			if e, ok := ev.(*fetch.EventAuthRequired); ok {
				go func() {
					_ = chromedp.Run(ctx, fetch.ContinueWithAuth(
						e.RequestID,
						&fetch.AuthChallengeResponse{
							Response: fetch.AuthChallengeResponseResponseProvideCredentials,
							Username: opts.ProxyUser,
							Password: opts.ProxyPass,
						},
					))
				}()
			}
		})
	}

	b := &Browser{
		allocCtx: allocCtx,
		ctx:      ctx,
		cancel: func() {
			cancel()
			allocCancel()
		},
	}
	return b, nil
}

func (b *Browser) Context() context.Context { return b.ctx }

func (b *Browser) Close() { b.cancel() }

func (b *Browser) AttachCollector(c *Collector) {
	b.collector = c
	c.Attach(b.ctx)
}
