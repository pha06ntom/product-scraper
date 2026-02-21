package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

func (b *Browser) CollectCategory(ctx context.Context, categoryURL string, scrolls int) error {
	return chromedp.Run(ctx,
		chromedp.Navigate(categoryURL),
		chromedp.Sleep(5*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < scrolls; i++ {
				_ = chromedp.Run(ctx, chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight);`, nil))
				time.Sleep(1200 * time.Millisecond)
			}
			return nil
		}),
		chromedp.Sleep(3*time.Second),
	)
}
