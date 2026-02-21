package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// Выбор адреса доставки обязателен,
// так как сайт возвращает разные цены и доступность
// в зависимости от выбранного магазина.
//
// Используются "мягкие" XPath-поиски по текстам и placeholder,
// чтобы минимизировать зависимость от конкретной разметки.
func (b *Browser) SelectAddress(ctx context.Context, address string) error {
	steps := []chromedp.Action{
		chromedp.Navigate("https://lenta.com"),
		chromedp.Sleep(3 * time.Second),

		clickFirstMatch([]string{
			`//button[contains(., 'Адрес')]`,
			`//button[contains(., 'Доставка')]`,
			`//a[contains(., 'Адрес')]`,
			`//a[contains(., 'Доставка')]`,
		}),
		chromedp.Sleep(200 * time.Millisecond),

		chromedp.SetValue(`//input[contains(@placeholder,'Адрес') or contains(@aria-label,'Адрес') or contains(@name,'address') or contains(@placeholder,'улиц')]`, ""),
		chromedp.SendKeys(`//input[contains(@placeholder,'Адрес') or contains(@aria-label,'Адрес') or contains(@name,'address') or contains(@placeholder,'улиц')]`, address),
		chromedp.Sleep(1500 * time.Millisecond),

		clickFirstMatch([]string{
			`(//li[contains(@class,'suggest') or contains(@class,'dropdown') or contains(@class,'option')])[1]`,
			`(//div[contains(@class,'suggest') or contains(@class,'dropdown') or contains(@class,'option')])[1]`,
			`(//button[contains(@class,'suggest') or contains(@class,'dropdown') or contains(@class,'option')])[1]`,
		}),
		chromedp.Sleep(2 * time.Second),

		tryClickAny([]string{
			`//button[contains(., 'Подтвердить')]`,
			`//button[contains(., 'Сохранить')]`,
			`//button[contains(., 'Выбрать')]`,
			`//button[contains(., 'Готово')]`,
		}),
		chromedp.Sleep(2 * time.Second),
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		lastErr = chromedp.Run(ctx, steps...)
		if lastErr == nil {
			return nil
		}
		time.Sleep(1500 * time.Millisecond)
	}
	return fmt.Errorf("address selection failed: %w", lastErr)
}

func clickFirstMatch(xpaths []string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, xp := range xpaths {
			var nodes []*cdp.Node
			if err := chromedp.Run(ctx, chromedp.Nodes(xp, &nodes, chromedp.BySearch)); err == nil && len(nodes) > 0 {
				return chromedp.Run(ctx, chromedp.Click(xp, chromedp.BySearch))
			}
		}
		return nil
	})
}

func focusFirstMatch(xpaths []string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, xp := range xpaths {
			var nodes []*cdp.Node
			if err := chromedp.Run(ctx, chromedp.Nodes(xp, &nodes, chromedp.BySearch)); err == nil && len(nodes) > 0 {
				return chromedp.Run(ctx, chromedp.Focus(xp, chromedp.BySearch))
			}
		}
		return nil
	})
}

func tryClickAny(xpaths []string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		for _, xp := range xpaths {
			var nodes []*cdp.Node
			if err := chromedp.Run(ctx, chromedp.Nodes(xp, &nodes, chromedp.BySearch)); err == nil && len(nodes) > 0 {
				_ = chromedp.Run(ctx, chromedp.Click(xp, chromedp.BySearch))
				return nil
			}
		}
		return nil
	})
}
