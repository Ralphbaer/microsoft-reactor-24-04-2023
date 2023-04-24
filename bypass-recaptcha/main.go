package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"time"

	api2captcha "github.com/2captcha/2captcha-go"
	"github.com/joho/godotenv"

	"github.com/chromedp/chromedp"
)

func wait(sel string) chromedp.ActionFunc {
	return run(1*time.Second, chromedp.WaitReady(sel))
}

func run(timeout time.Duration, task chromedp.Action) chromedp.ActionFunc {
	return runFunc(timeout, task.Do)
}

func runFunc(timeout time.Duration, task chromedp.ActionFunc) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		return task.Do(ctx)
	}
}

func solveReCaptcha(client *api2captcha.Client, targetURL, dataSiteKey string) (string, error) {
	c := api2captcha.ReCaptcha{
		SiteKey:   dataSiteKey,
		Url:       targetURL,
		Invisible: true,
		Action:    "verify",
	}

	return client.Solve(c.ToRequest())
}

func recaptchaDemoActions(client *api2captcha.Client) []chromedp.Action {
	const targetURL string = "https://www.google.com/recaptcha/api2/demo"
	var siteKey string
	var siteKeyOk bool

	return []chromedp.Action{
		run(5*time.Second, chromedp.Navigate(targetURL)),
		wait(`[data-sitekey]`),
		wait(`#g-recaptcha-response`),
		wait(`#recaptcha-demo-submit`),
		run(time.Second, chromedp.AttributeValue(`[data-sitekey]`, "data-sitekey", &siteKey, &siteKeyOk)),
		runFunc(5*time.Minute, func(ctx context.Context) error {
			if !siteKeyOk {
				return errors.New("missing data-sitekey")
			}

			token, err := solveReCaptcha(client, targetURL, siteKey)
			if err != nil {
				return err
			}

			return chromedp.
				SetJavascriptAttribute(`#g-recaptcha-response`, "innerText", token).
				Do(ctx)
		}),
		chromedp.Click(`#recaptcha-demo-submit`),
		wait(`.recaptcha-success`),
	}
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	client := api2captcha.NewClient(os.Getenv("API_KEY"))
	actions := recaptchaDemoActions(client)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.WindowSize(1366, 768),
		chromedp.Flag("headless", false),
		chromedp.Flag("incognito", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", true),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	start := time.Now()
	err := chromedp.Run(ctx, actions...)
	end := time.Since(start)

	if err != nil {
		log.Println("bypass recaptcha failed:", err, end)
		return
	}

	var buf []byte
	if err := chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf)); err != nil {
		log.Println("failed to capture screenshot", err)
	}

	// Save the screenshot as a PNG file
	if err := ioutil.WriteFile("screenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}

	log.Println("bypass recaptcha success", end)
}
