package gui

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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
	// ── Navigation Rail ───────────────────────────────────────────────────────
	nav := container.NewVBox(
		widget.NewButtonWithIcon("", theme.HomeIcon(), func() {}),
		widget.NewButtonWithIcon("", theme.HistoryIcon(), pa.showHistory),
		widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.HelpIcon(), func() {}),
	)
	navRail := container.NewPadded(nav)

	// ── Dashboard Header ──────────────────────────────────────────────────────
	title := widget.NewLabelWithStyle("Dashboard", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	title.Alignment = fyne.TextAlignLeading
	
	totalLabel := binding.NewString()
	totalLabel.Set("0")
	activeLabel := binding.NewString()
	activeLabel.Set("0")
	memLabel := binding.NewString()
	memLabel.Set("0 MB")

	createCard := func(title, icon string, data binding.String) fyne.CanvasObject {
		lbl := widget.NewLabelWithData(data)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		return container.NewPadded(container.NewVBox(
			widget.NewLabel(title),
			lbl,
		))
	}

	stats := container.NewGridWithColumns(3,
		createCard("TOTAL PORTS", "home", totalLabel),
		createCard("ACTIVE SERVICES", "check", activeLabel),
		createCard("TOTAL MEMORY", "info", memLabel),
	)

	// ── Search & List ─────────────────────────────────────────────────────────
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search ports or services...")
	searchEntry.ActionItem = widget.NewIcon(theme.SearchIcon())
	searchEntry.OnChanged = func(s string) {
		pa.searchQuery.Set(s)
		pa.filterPorts()
	}

	list := widget.NewListWithData(
		pa.ports,
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.HelpIcon())
			name := widget.NewLabelWithStyle("Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			service := widget.NewLabel("Service")
			port := widget.NewLabel(":8080")
			
			killBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
			killBtn.Importance = widget.DangerImportance
			
			return container.NewBorder(nil, nil, 
				container.NewHBox(icon, name), 
				killBtn,
				container.NewHBox(service, port),
			)
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			val, _ := i.(binding.Untyped).Get()
			p := val.(scanner.PortEntry)

			border := o.(*fyne.Container)
			leftBox := border.Objects[1].(*fyne.Container)
			icon := leftBox.Objects[0].(*widget.Icon)
			name := leftBox.Objects[1].(*widget.Label)
			
			centerBox := border.Objects[0].(*fyne.Container)
			service := centerBox.Objects[0].(*widget.Label)
			port := centerBox.Objects[1].(*widget.Label)
			
			killBtn := border.Objects[2].(*widget.Button)

			icon.SetResource(GetServiceIcon(p.Service))
			name.SetText(p.Name)
			service.SetText(p.Service)
			port.SetText(fmt.Sprintf(":%d", p.Port))
			
			killBtn.OnTapped = func() {
				pa.confirmKill(p)
			}
		},
	)

	mainContent := container.NewBorder(
		container.NewVBox(
			container.NewPadded(title),
			container.NewPadded(stats),
			widget.NewSeparator(),
			container.NewPadded(searchEntry),
		),
		nil, nil, nil,
		container.NewPadded(list),
	)

	// ── Final Layout ──────────────────────────────────────────────────────────
	pa.Window.SetContent(container.NewBorder(nil, nil, navRail, nil, mainContent))

	// ── Background Updates ────────────────────────────────────────────────────
	go func() {
		for {
			totalLabel.Set(fmt.Sprintf("%d", len(pa.allPorts)))
			// Count unique PIDs for active services
			pids := make(map[int32]bool)
			for _, p := range pa.allPorts { pids[p.PID] = true }
			activeLabel.Set(fmt.Sprintf("%d", len(pids)))
			
			pa.Window.Content().Refresh()
			time.Sleep(2 * time.Second)
		}
	}()
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
