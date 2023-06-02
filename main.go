package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"io/ioutil"
	"strings"
)

type config struct {
	EditorWidget  *widget.Entry
	PreviewWidget *widget.RichText
	CurrentFile   fyne.URI
	SaveMenuItem  *fyne.MenuItem
}

var cfg config

func main() {
	a := app.New()
	a.Settings().SetTheme(&myTheme{})
	wind := a.NewWindow("MarkdownEditor")

	edit, preview := cfg.MakeUI()
	wind.SetContent(container.NewHSplit(edit, preview))
	wind.Resize(fyne.Size{
		Width:  800,
		Height: 500,
	})

	cfg.MakeMenu(wind)

	wind.CenterOnScreen()
	wind.ShowAndRun()
}

func (c *config) MakeUI() (*widget.Entry, *widget.RichText) {
	edit := widget.NewMultiLineEntry()
	preview := widget.NewRichTextFromMarkdown("")

	c.EditorWidget = edit
	c.PreviewWidget = preview

	edit.OnChanged = preview.ParseMarkdown
	return edit, preview
}

func (c *config) MakeMenu(window fyne.Window) {
	openMenuItem := fyne.NewMenuItem("Open...", c.openFunc(window))

	saveMenuItem := fyne.NewMenuItem("Save", c.saveFunc(window))
	saveMenuItem.Disabled = true
	c.SaveMenuItem = saveMenuItem

	saveAsMenuItem := fyne.NewMenuItem("Save as...", c.saveAsFunc(window))

	mainMenu := fyne.NewMenu("File", openMenuItem, saveMenuItem, saveAsMenuItem)
	menu := fyne.NewMainMenu(mainMenu)
	window.SetMainMenu(menu)
}

var filter = storage.NewExtensionFileFilter([]string{".md", ".MD"})

func (c *config) openFunc(window fyne.Window) func() {
	return func() {
		readDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			defer func(reader fyne.URIReadCloser) {
				err = reader.Close()
				if err != nil {
					dialog.ShowError(err, window)
					return
				}
			}(reader)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			if reader == nil {
				//点击取消
				return
			}

			data, err := ioutil.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			c.EditorWidget.SetText(string(data))
			c.CurrentFile = reader.URI()
			window.SetTitle(window.Title() + " - " + reader.URI().Name())
			c.SaveMenuItem.Disabled = false
		}, window)

		readDialog.SetFilter(filter)
		readDialog.Show()
	}
}

func (c *config) saveFunc(win fyne.Window) func() {
	return func() {
		if c.CurrentFile != nil {
			writer, err := storage.Writer(c.CurrentFile)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			defer func(writer fyne.URIWriteCloser) {
				err = writer.Close()
				if err != nil {
					dialog.ShowError(err, win)
					return
				}
			}(writer)

			_, err = writer.Write([]byte(c.EditorWidget.Text))
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
		}
	}
}

func (c *config) saveAsFunc(win fyne.Window) func() {
	return func() {
		saveAsDialog := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
			defer func(write fyne.URIWriteCloser) {
				err = write.Close()
				if err != nil {
					dialog.ShowError(err, win)
					return
				}
			}(write)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if write == nil {
				//点了取消，write里没东西
				return
			}

			dialog.ShowInformation("test", write.URI().Extension(), win)
			if !strings.HasSuffix(strings.ToLower(write.URI().String()), ".md") {
				dialog.ShowInformation("Error", "File name should end with .md", win)
				return
			}

			_, err = write.Write([]byte(c.EditorWidget.Text))
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			c.CurrentFile = write.URI()
			win.SetTitle(win.Title() + " - " + write.URI().Name())
			c.SaveMenuItem.Disabled = false
		}, win)

		saveAsDialog.SetFileName("untitled.md")
		saveAsDialog.Show()
	}
}
