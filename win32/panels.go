package win32

type ControllerGroup struct {
	controllers      []Controller
	clickControllers map[int]ClickController
	// TODO: content per generics
}

func NewControllerGroup(controller ...Controller) *ControllerGroup {
	clickController := make(map[int]ClickController)
	for _, c := range controller {
		if cc, ok := c.(ClickController); ok {
			clickController[c.ID()] = cc
		}
	}
	return &ControllerGroup{
		controllers:      controller,
		clickControllers: clickController,
	}
}
func (cg *ControllerGroup) Controller() []Controller {
	return cg.controllers
}
func (cg *ControllerGroup) ClickController(id int) ClickController {
	return cg.clickControllers[id]
}

type PanelGroup interface {
	Controller() []Controller
	ClickController(id int) ClickController
	Resize(tabHeight, width, height int32)

	SaveState()
	SetState(data any)
}

type Panels struct {
	panels    map[PanelGroupName]PanelGroup
	active    PanelGroupName
	tabHeight int32
	width     int32
	height    int32
}

func NewPanels(tabHeight, width, height int32) *Panels {
	return &Panels{
		panels:    make(map[PanelGroupName]PanelGroup),
		tabHeight: tabHeight,
		width:     width,
		height:    height,
	}
}

type PanelGroupName string

func (p *Panels) Add(name PanelGroupName, pg PanelGroup) {
	p.panels[name] = pg
}

func (p *Panels) Show(panel PanelGroupName) {
	p.active = panel
	for name, pg := range p.panels {
		if name == panel {
			pg.Resize(p.tabHeight, p.width, p.height)
			for _, ctrl := range pg.Controller() {
				ctrl.Show()
			}
		} else {
			for _, ctrl := range pg.Controller() {
				ctrl.Hide()
			}
		}
	}
}

func (p *Panels) Resize(width, height int32) {
	p.width = width
	p.height = height
	p.panels[p.active].Resize(p.tabHeight, width, height)
}

func (p *Panels) get(panel PanelGroupName) PanelGroup {
	return p.panels[panel]
}

func (p *Panels) HandleCommand(id int, notifyCode int) {
	if cc := p.panels[p.active].ClickController(id); cc != nil {
		switch notifyCode {
		case BN_CLICKED:
			cc.OnClick()
		case LBN_DBLCLK:
			cc.OnDoubleClick()
		}
	}
}
