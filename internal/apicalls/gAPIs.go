package apicalls

import (
	"fmt"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/forms/v1"
)

type Caller struct {
	DriveService *drive.Service
	FormsService *forms.Service
}

func NewCaller(driveService *drive.Service, formsService *forms.Service) *Caller {
	return &Caller{
		DriveService: driveService,
		FormsService: formsService,
	}
}

var DrivePageToken string

func (p *Caller) DriveChanges() (*drive.ChangeList, error) {

	if DrivePageToken == "" {
		startToken, err := p.DriveService.Changes.GetStartPageToken().Do()
		if err != nil {
			return nil, fmt.Errorf("unable to get the start page token : %v", err)
		}
		DrivePageToken = startToken.StartPageToken
	}
	currentList, err := p.DriveService.Changes.List(DrivePageToken).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get list of changes in GDrive : %v", err)
	}
	DrivePageToken = currentList.NewStartPageToken

	return currentList, nil 
}

func (p *Caller) GetFormMetadata(formID string) (*forms.Form, error) {

	formData, err := p.FormsService.Forms.Get(formID).Fields("responderUri", "formId").Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get form metadata : %v", err)
	}

	return formData, nil
}

func (p *Caller) GetCompleteForm(formID string) (*forms.Form, error) {

	formData, err := p.FormsService.Forms.Get(formID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get form : %v", err)
	}

	return formData, nil
}
