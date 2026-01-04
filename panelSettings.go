package main

import (
	"fmt"

	"hoermi.com/rest-test/win32"
)

type settingsPanelGroup struct {
	*win32.ControllerGroup
	// Settings Panel controls
	certInput       *win32.Control
	keyInput        *win32.Control
	caInput         *win32.Control
	skipVerifyChk   *win32.CheckBoxControl
	certBtn         *win32.ButtonControl
	keyBtn          *win32.ButtonControl
	caBtn           *win32.ButtonControl
	certLabel       *win32.Control
	keyLabel        *win32.Control
	caLabel         *win32.Control
	settingsTitle   *win32.Control
	saveSettingsBtn *win32.ButtonControl

	content *SettingsTabContent
}

func (s *settingsPanelGroup) Resize(tabHeight, width, height int32) {
	y := tabHeight + layoutPadding

	s.settingsTitle.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight+4)
	y += layoutLabelHeight + layoutPadding

	s.certLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	s.certInput.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	s.certBtn.MoveWindow(layoutPadding+layoutColumnWidth+layoutPadding, y, layoutButtonWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	s.keyLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	s.keyInput.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	s.keyBtn.MoveWindow(layoutPadding+layoutColumnWidth+layoutPadding, y, layoutButtonWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	s.caLabel.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutLabelHeight)
	y += layoutLabelHeight + layoutPadding
	s.caInput.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	s.caBtn.MoveWindow(layoutPadding+layoutColumnWidth+layoutPadding, y, layoutButtonWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding

	s.skipVerifyChk.MoveWindow(layoutPadding, y, layoutColumnWidth, layoutInputHeight)
	y += layoutInputHeight + layoutPadding
	s.saveSettingsBtn.MoveWindow(layoutPadding, y, layoutButtonWidth, layoutInputHeight)
}

func (s *settingsPanelGroup) SaveState() {
	s.content.Settings.Certificate.CertFile = s.certInput.GetText()
	s.content.Settings.Certificate.KeyFile = s.keyInput.GetText()
	s.content.Settings.Certificate.CACertFile = s.caInput.GetText()
	s.content.Settings.Certificate.SkipVerify = s.skipVerifyChk.IsChecked()
}

func (s *settingsPanelGroup) SetState(data any) {
	if content, ok := data.(*SettingsTabContent); ok {
		s.content = content
		s.certInput.SetText(content.Settings.Certificate.CertFile)
		s.keyInput.SetText(content.Settings.Certificate.KeyFile)
		s.caInput.SetText(content.Settings.Certificate.CACertFile)
		s.skipVerifyChk.SetChecked(content.Settings.Certificate.SkipVerify)
	}
}

func createSettingsPanel(factory win32.ControlFactory) *settingsPanelGroup {
	group := &settingsPanelGroup{
		settingsTitle: factory.CreateLabel("Global Settings"),
		certLabel:     factory.CreateLabel("Client Certificate (PEM)"),
		certInput:     factory.CreateInput(),
		keyLabel:      factory.CreateLabel("Private Key (PEM)"),
		keyInput:      factory.CreateInput(),
		caLabel:       factory.CreateLabel("CA Bundle (optional)"),
		caInput:       factory.CreateInput(),
		skipVerifyChk: factory.CreateCheckbox("Skip TLS Verification (insecure)"),
	}
	group.certBtn = factory.CreateButton("...", func() {
		if path, ok := factory.OpenFileDialog("Select Certificate", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			group.certInput.SetText(path)
		}

	})
	group.keyBtn = factory.CreateButton("...", func() {
		if path, ok := factory.OpenFileDialog("Select Private Key", "PEM Files (*.pem;*.key)|*.pem;*.key|All Files (*.*)|*.*|", "pem"); ok {
			group.keyInput.SetText(path)
		}

	})
	group.caBtn = factory.CreateButton("...", func() {
		if path, ok := factory.OpenFileDialog("Select CA Bundle", "PEM Files (*.pem;*.crt)|*.pem;*.crt|All Files (*.*)|*.*|", "pem"); ok {
			group.caInput.SetText(path)
		}
	})
	group.saveSettingsBtn = factory.CreateButton("Save Settings", func() {
		group.SaveState()
		if err := group.content.Settings.save(); err != nil {
			factory.MessageBox("Error", fmt.Sprintf("Error saving settings: %v", err))
		}
	})
	group.ControllerGroup = win32.NewControllerGroup(
		group.certInput, group.keyInput, group.caInput,
		group.skipVerifyChk, group.certBtn, group.keyBtn, group.caBtn,
		group.settingsTitle, group.certLabel, group.keyLabel, group.caLabel,
		group.saveSettingsBtn,
	)
	return group
}
