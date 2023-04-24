package main

import (
	"context"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false), // The --disable-gpu option is a temporary workaround for a few bugs
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("incognito", true),
	)

	// Create a new ExecAllocator to customize how Chrome will be started
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create a new context with the customized allocator and logging enabled
	ctx, ctxCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer ctxCancel()

	// Search for a video on Youtube and retrieve its link
	var videoLink string
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.youtube.com"),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.SendKeys(`input#search`, "Golang SP: Go is Back! @Microsoft Reactor"),
		chromedp.Submit("input#search"),
		chromedp.WaitVisible(`a#video-title`, chromedp.ByQuery),
		chromedp.AttributeValue(`a#video-title`, "href", &videoLink, nil),
	); err != nil {
		log.Fatal(err)
	}

	// Capture a screenshot of the video page while it is playing
	var buf []byte
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.youtube.com"+videoLink),
		chromedp.WaitVisible(`video`, chromedp.ByQuery),
		chromedp.Click(`video`, chromedp.ByQuery),
		chromedp.Sleep(10*time.Second), // Wait for video to start playing
		chromedp.CaptureScreenshot(&buf),
	); err != nil {
		log.Fatal(err)
	}

	// Save the screenshot as a PNG file
	if err := ioutil.WriteFile("screenshot.png", buf, 0o644); err != nil {
		log.Fatal(err)
	}
}
