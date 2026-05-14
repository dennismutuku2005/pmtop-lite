package gui

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
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
	// Sidebar
	sidebar := container.NewVBox(
		widget.NewLabelWithStyle("FILTER", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Status"),
		widget.NewSelect([]string{"Active", "Listen", "Established"}, func(s string) {}),
		widget.NewLabel("Protocol"),
		widget.NewCheckGroup([]string{"TCP", "UDP"}, func(s []string) {}),
		container.NewSpacer(),
		widget.NewButtonWithIcon("History", theme.HistoryIcon(), pa.showHistory),
	)
	sidebarScroll := container.NewVScroll(container.NewPadded(sidebar))
	sidebarScroll.SetMinSize(fyne.NewSize(200, 0))

	// Header Cards
	totalLabel := widget.NewLabel("0")
	activeLabel := widget.NewLabel("0")
	memLabel := widget.NewLabel("0 MB")

	createCard := func(title string, label *widget.Label) fyne.CanvasObject {
		return container.NewPadded(container.NewVBox(
			widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			label,
		))
	}

	headerCards := container.NewGridWithColumns(3,
		createCard("TOTAL PORTS", totalLabel),
		createCard("ACTIVE APPS", activeLabel),
		createCard("MEM USAGE", memLabel),
	)

	// Search and List
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search vehicle (process)...")
	searchEntry.OnChanged = func(s string) {
		pa.searchQuery.Set(s)
		pa.filterPorts()
	}

	list := widget.NewListWithData(
		pa.ports,
		func() fyne.CanvasObject {
			name := widget.NewLabelWithStyle("Name", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			service := widget.NewLabel("Service")
			port := widget.NewLabel(":8080")
			status := widget.NewLabel("Active")
			status.Importance = widget.HighImportance

			return container.NewGridWithColumns(4, name, service, port, status)
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			val, _ := i.(binding.Untyped).Get()
			p := val.(scanner.PortEntry)

			grid := o.(*fyne.Container)
			grid.Objects[0].(*widget.Label).SetText(p.Name)
			grid.Objects[1].(*widget.Label).SetText(p.Service)
			grid.Objects[2].(*widget.Label).SetText(fmt.Sprintf(":%d", p.Port))
			grid.Objects[3].(*widget.Label).SetText(p.State)
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		val, _ := pa.ports.GetValue(id)
		pa.showDetail(val.(scanner.PortEntry))
		list.Unselect(id)
	}

	mainContent := container.NewBorder(
		container.NewVBox(
			container.NewPadded(headerCards),
			widget.NewSeparator(),
			container.NewPadded(searchEntry),
		),
		nil, nil, nil,
		container.NewPadded(list),
	)

	// Final Layout
	pa.Window.SetContent(container.NewBorder(nil, nil, sidebarScroll, nil, mainContent))

	// Update stats loop
	go func() {
		for {
			totalLabel.SetText(fmt.Sprintf("%d", len(pa.allPorts)))
			// Simple counts for demo
			activeLabel.SetText(fmt.Sprintf("%d", len(pa.allPorts))) 
			pa.Window.Content().Refresh()
			pa.Window.Canvas().Refresh(totalLabel)
		}
	}()
}

func (pa *PortApp) filterPorts() {
	query, _ := pa.searchQuery.Get()
	query = strings.ToLower(query)

	filtered := make([]interface{}, 0)
	for _, p := range pa.allPorts {
		if query == "" || 
			strings.Contains(strings.ToLower(p.Name), query) || 
			strings.Contains(fmt.Sprintf("%d", p.Port), query) {
			filtered = append(filtered, p)
		}
	}
	pa.ports.Set(filtered)
}

func (pa *PortApp) showDetail(p scanner.PortEntry) {
	d := pa.App.NewWindow(fmt.Sprintf("Vehicle Details: %d", p.Port))
	d.SetContent(widget.NewLabel(fmt.Sprintf("Details for %s on port %d", p.Name, p.Port)))
	d.Resize(fyne.NewSize(300, 200))
	d.Show()
}

func (pa *PortApp) showHistory() {
	// ... (history implementation)
}

func (pa *PortApp) parseURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
