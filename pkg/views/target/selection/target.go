// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package selection

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

var NewTargetIdentifier = "<NEW_TARGET>"

func generateTargetList(targets []apiclient.TargetDTO, isMultipleSelect bool, action string) []list.Item {
	// Initialize an empty list of items.
	items := []list.Item{}

	// Populate items with titles and descriptions from targets.
	for _, target := range targets {
		var providerName string

		if target.ProviderInfo.Label != nil {
			providerName = *target.ProviderInfo.Label
		} else {
			providerName = target.ProviderInfo.Name
		}

		newItem := item[apiclient.TargetDTO]{
			title:          target.Name,
			id:             target.Id,
			desc:           providerName,
			target:         &target,
			choiceProperty: target,
		}

		if isMultipleSelect {
			newItem.isMultipleSelect = true
			newItem.action = action
		}

		items = append(items, newItem)
	}

	return items
}

func getTargetProgramEssentials(modelTitle string, withNewTarget bool, actionVerb string, targets []apiclient.TargetDTO, footerText string, isMultipleSelect bool) tea.Model {

	items := generateTargetList(targets, isMultipleSelect, actionVerb)

	if withNewTarget {
		newItem := item[apiclient.TargetDTO]{title: "+ Create a new target", desc: "", choiceProperty: apiclient.TargetDTO{
			Name: NewTargetIdentifier,
		}}
		items = append(items, newItem)
	}

	d := ItemDelegate[apiclient.TargetDTO]{}

	l := list.New(items, d, 0, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(views.Green)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(views.Green)

	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(views.Green)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(views.Green)

	m := model[apiclient.TargetDTO]{list: l}

	m.list.Title = views.GetStyledMainTitle(modelTitle + actionVerb)
	m.list.Styles.Title = lipgloss.NewStyle().Foreground(views.Green).Bold(true)
	m.footer = footerText

	p, err := tea.NewProgram(m, tea.WithAltScreen()).Run()

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return p
}

func selectTargetPrompt(targets []apiclient.TargetDTO, withNewTarget bool, actionVerb string, choiceChan chan<- *apiclient.TargetDTO) {
	p := getTargetProgramEssentials("Select a Target To ", withNewTarget, actionVerb, targets, "", false)
	if m, ok := p.(model[apiclient.TargetDTO]); ok && m.choice != nil {
		choiceChan <- m.choice
	} else {
		choiceChan <- nil
	}
}

func GetTargetFromPrompt(targets []apiclient.TargetDTO, withNewTarget bool, actionVerb string) *apiclient.TargetDTO {
	choiceChan := make(chan *apiclient.TargetDTO)

	go selectTargetPrompt(targets, withNewTarget, actionVerb, choiceChan)

	return <-choiceChan
}

func selectTargetsFromPrompt(targets []apiclient.TargetDTO, actionVerb string, choiceChan chan<- []*apiclient.TargetDTO) {
	footerText := lipgloss.NewStyle().Bold(true).PaddingLeft(2).Render(fmt.Sprintf("\n\nPress 'x' to mark a target.\nPress 'enter' to %s the current/marked targets.", actionVerb))
	p := getTargetProgramEssentials("Select Targets To ", false, actionVerb, targets, footerText, true)

	m, ok := p.(model[apiclient.TargetDTO])
	if ok && m.choices != nil {
		choiceChan <- m.choices
	} else if ok && m.choice != nil {
		choiceChan <- []*apiclient.TargetDTO{m.choice}
	} else {
		choiceChan <- nil
	}
}

func GetTargetsFromPrompt(targets []apiclient.TargetDTO, actionVerb string) []*apiclient.TargetDTO {
	choiceChan := make(chan []*apiclient.TargetDTO)

	go selectTargetsFromPrompt(targets, actionVerb, choiceChan)

	return <-choiceChan
}