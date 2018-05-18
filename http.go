package communication

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/interactive-solutions/go-communication/internal"
)

type HttpHandler struct {
	app *application
}

type collectionMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

func (h *HttpHandler) TestTemplate(w http.ResponseWriter, r *http.Request) {

	body := &internal.TestTemplateRequest{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		http.Error(w, "Failed to parse incoming json", 400)
		return
	}

	split := strings.SplitN(body.Id, ":", 2)
	if len(split) != 2 {
		http.Error(w, "Invalid id provided, templateId:locale expected", 400)
		return
	}

	template, err := h.app.templateRepo.Get(split[1], split[0])
	if err != nil {
		if err == TemplateNotFoundErr {
			http.Error(w, "Template not found", 404)
			return
		}

		http.Error(w, "Failed to retrieve template", 500)
		return
	}

	switch body.Type {
	case "sms":
		h.app.SendSms(template.TemplateId, template.Locale, body.Target, "", template.Parameters)

	case "email":
		h.app.SendEmail(template.TemplateId, template.Locale, body.Target, "", template.Parameters)

	default:
		http.Error(w, fmt.Sprintf("Unsupported type %s", body.Type), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HttpHandler) GetAllTemplates(w http.ResponseWriter, r *http.Request) {

	criteria := PopulateTemplateCriteria(r)

	templates, count, err := h.app.templateRepo.Matching(criteria)
	if err != nil {
		http.Error(w, "Failed to retrieve templates", 500)
		return
	}

	payload := struct {
		Data []Template     `json:"data"`
		Meta collectionMeta `json:"meta"`
	}{
		Data: templates,
		Meta: collectionMeta{
			Total:  count,
			Limit:  criteria.Limit,
			Offset: criteria.Offset,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to convert to json", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *HttpHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "Route id var", 400)
		return
	}

	split := strings.SplitN(id, ":", 2)
	if len(split) != 2 {
		http.Error(w, "Invalid id provided, templateId:locale expected", 400)
		return
	}

	template, err := h.app.templateRepo.Get(split[1], split[0])
	if err != nil {
		if err == TemplateNotFoundErr {
			http.Error(w, "Template not found", 404)
			return
		}

		http.Error(w, "Failed to retrieve template", 500)
		return
	}

	data, err := json.Marshal(template)
	if err != nil {
		http.Error(w, "Failed to convert template to json", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *HttpHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {

	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "Route id var", 400)
		return
	}

	split := strings.SplitN(id, ":", 2)
	if len(split) != 2 {
		http.Error(w, "Invalid id provided, templateId:locale expected", 400)
		return
	}

	template, err := h.app.templateRepo.Get(split[1], split[0])
	if err != nil {
		if err == TemplateNotFoundErr {
			http.Error(w, "Template not found", 404)
			return
		}

		http.Error(w, "Failed to retrieve template", 500)
		return
	}

	body := &internal.UpdateTemplateRequest{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		http.Error(w, "Failed to parse incoming json", 400)
		return
	}

	template.Subject = body.Subject
	template.TextBody = body.TextBody
	template.HtmlBody = body.HtmlBody
	template.UpdateParameters = body.UpdateParameters
	template.Enabled = body.Enabled

	// Check if we have a html to text converter if the text body was not provided
	if template.TextBody == "" && h.app.htmlToTextConverter != nil {
		template.TextBody = h.app.htmlToTextConverter(template.TextBody)
	}

	if _, _, _, err := h.app.Render(template, &Job{Params: template.Parameters}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template with error: %s", err.Error()), 422)
		return
	}

	if err := h.app.templateRepo.Update(&template); err != nil {
		http.Error(w, "Failed to update template", 500)
		return
	}

	data, err := json.Marshal(template)
	if err != nil {
		http.Error(w, "Failed to convert template to json", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *HttpHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "Route id var", 400)
		return
	}

	split := strings.SplitN(id, ":", 2)
	if len(split) != 2 {
		http.Error(w, "Invalid id provided, templateId:locale expected", 400)
		return
	}

	template, err := h.app.templateRepo.Get(split[1], split[0])
	if err != nil {
		if err == TemplateNotFoundErr {
			http.Error(w, "Template not found", 404)
			return
		}

		http.Error(w, "Failed to retrieve template", 500)
		return
	}

	if err := h.app.templateRepo.Delete(&template); err != nil {
		http.Error(w, "Failed to delete template", 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
