package gui

import (
	"context"
	"fmt"
	"image/color"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/dennismutuku2005/pmtop-lite/pkg/scanner"
	"github.com/dennismutuku2005/pmtop-lite/pkg/state"
)

type PortApp struct {
	App    fyne.App
	Window fyne.Window

	mgr    *state.Manager
	cancel context.CancelFunc

	ports       binding.UntypedList
	searchQuery binding.String
	showAll     binding.Bool

	allPorts []scanner.PortEntry
}

func NewPortApp() *PortApp {
	a := app.NewWithID("com.pmtop.gui")
	a.Settings().SetTheme(&ModernLightTheme{})

	w := a.NewWindow("pmtop — Fleet Port Monitor")
	w.Resize(fyne.NewSize(1000, 700))

	pa := &PortApp{
		App:         a,
		Window:      w,
		ports:       binding.NewUntypedList(),
		searchQuery: binding.NewString(),
		showAll:     binding.NewBool(),
	}

	pa.mgr = state.New(false)
	return pa
}

func (pa *PortApp) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	pa.cancel = cancel
	go pa.mgr.Run(ctx)

	go func() {
		for msg := range pa.mgr.Updates() {
			pa.allPorts = msg.Ports
			pa.filterPorts()
		}
	}()

	pa.setupUI()
	pa.Window.ShowAndRun()
}

func (pa *PortApp) setupUI() {
	// ── Tabs ──────────────────────────────────────────────────────────────────
	dashboard := pa.buildDashboard()
	settings := pa.buildSettings()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Dashboard", theme.HomeIcon(), dashboard),
		container.NewTabItemWithIcon("Settings", theme.SettingsIcon(), settings),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	pa.Window.SetContent(tabs)
}

func (pa *PortApp) buildDashboard() fyne.CanvasObject {
	// ── Stats Header ──────────────────────────────────────────────────────────
	totalLabel := binding.NewString()
	totalLabel.Set("0")
	activeLabel := binding.NewString()
	activeLabel.Set("0")

	createStat := func(title string, data binding.String, color color.Color) fyne.CanvasObject {
		lbl := widget.NewLabelWithData(data)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		return container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			container.NewCenter(lbl),
		)
	}

	stats := container.NewGridWithColumns(2,
		createStat("TOTAL PORTS", totalLabel, theme.Color(theme.ColorNamePrimary, theme.VariantLight)),
		createStat("ACTIVE SERVICES", activeLabel, theme.Color(theme.ColorNameSuccess, theme.VariantLight)),
	)

	// ── Port List ─────────────────────────────────────────────────────────────
	list := widget.NewListWithData(
		pa.ports,
		func() fyne.CanvasObject {
			dot := canvas.NewCircle(color.NRGBA{R: 0, G: 255, B: 0, A: 255})
			dot.Resize(fyne.NewSize(10, 10))
			
			name := widget.NewLabelWithStyle("Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			port := widget.NewLabel(":8080")
			mem := widget.NewLabel("0 MB")
			mem.Importance = widget.LowImportance
			
			killBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			killBtn.Importance = widget.DangerImportance
			
			return container.NewBorder(nil, nil, 
				container.NewHBox(container.NewCenter(dot), name), 
				killBtn,
				container.NewHBox(port, widget.NewSeparator(), mem),
			)
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			val, _ := i.(binding.Untyped).Get()
			p := val.(scanner.PortEntry)

			border := o.(*fyne.Container)
			leftBox := border.Objects[1].(*fyne.Container)
			dot := leftBox.Objects[0].(*fyne.Container).Objects[0].(*canvas.Circle)
			name := leftBox.Objects[1].(*widget.Label)
			
			rightBox := border.Objects[0].(*fyne.Container)
			port := rightBox.Objects[0].(*widget.Label)
			mem := rightBox.Objects[2].(*widget.Label)
			
			killBtn := border.Objects[2].(*widget.Button)

			if p.Service == "Unknown" {
				dot.FillColor = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
			} else {
				dot.FillColor = color.NRGBA{R: 254, G: 74, B: 22, A: 255} // Brand Orange
			}
			dot.Refresh()

			name.SetText(p.Name)
			port.SetText(fmt.Sprintf(":%d", p.Port))
			mem.SetText(proc.FormatMem(p.MemoryMB))
			
			killBtn.OnTapped = func() {
				pa.confirmKill(p)
			}
		},
	)

	// ── Background Updates ────────────────────────────────────────────────────
	go func() {
		for {
			totalLabel.Set(fmt.Sprintf("%d", len(pa.allPorts)))
			pids := make(map[int32]bool)
			for _, p := range pa.allPorts { pids[p.PID] = true }
			activeLabel.Set(fmt.Sprintf("%d", len(pids)))
			time.Sleep(2 * time.Second)
		}
	}()

	return container.NewBorder(
		container.NewVBox(container.NewPadded(stats), widget.NewSeparator()),
		nil, nil, nil,
		container.NewPadded(list),
	)
}

func (pa *PortApp) buildSettings() fyne.CanvasObject {
	updateCheck := widget.NewCheck("Check for updates on startup", func(b bool) {})
	updateCheck.Checked = true
	
	return container.NewPadded(container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		updateCheck,
		container.NewSpacer(),
		widget.NewLabel("pmtop v0.1.0"),
		widget.NewButton("Check for Updates Now", func() {
			// Mock update check
			pa.Window.SetContent(container.NewCenter(widget.NewLabel("You are on the latest version!")))
			time.AfterFunc(2*time.Second, func() { pa.setupUI() })
		}),
	))
}


func (pa *PortApp) confirmKill(p scanner.PortEntry) {
	confirm := widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabelWithStyle("Confirm Close", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel(fmt.Sprintf("Are you sure you want to close %s on port %d?", p.Name, p.Port)),
			container.NewHBox(
				widget.NewButton("Cancel", func() { pa.Window.Canvas().Overlays().Remove(pa.Window.Canvas().Overlays().Top()) }),
				widget.NewButtonWithIcon("Close Process", theme.DeleteIcon(), func() {
					pa.mgr.Kill(p.PID)
					pa.Window.Canvas().Overlays().Remove(pa.Window.Canvas().Overlays().Top())
				}),
			),
		),
		pa.Window.Canvas(),
	)
	confirm.Show()
}

func (pa *PortApp) filterPorts() {
	query, _ := pa.searchQuery.Get()
	query = strings.ToLower(query)

	filtered := make([]interface{}, 0)
	for _, p := range pa.allPorts {
		if query == "" || 
			strings.Contains(strings.ToLower(p.Name), query) || 
			strings.Contains(fmt.Sprintf("%d", p.Port), query) ||
			strings.Contains(strings.ToLower(p.Service), query) {
			filtered = append(filtered, p)
		}
	}
	pa.ports.Set(filtered)
}

func (pa *PortApp) showDetail(p scanner.PortEntry) {
	// Detail view handled in main list via Quick Actions or Selection
}

func (pa *PortApp) showHistory() {
	// ... (history implementation)
}


func (pa *PortApp) parseURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
