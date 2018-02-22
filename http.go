package communication

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/interactive-solutions/go-communication/internal"
)

type HttpHandler struct {
	app *application
}

func (h *HttpHandler) GetAllTemplates(w http.ResponseWriter, r *http.Request) {

	templates, err := h.app.templateRepo.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve templates", 500)
		return
	}

	payload := struct {
		Data []Template `json:"data"`
	}{templates}

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

	template, err := h.app.templateRepo.Get(split[0], split[1])
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

	template, err := h.app.templateRepo.Get(split[0], split[1])
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

	template.UpdateParameters = body.UpdateParameters
	template.Enabled = body.Enabled
	template.Subject = body.Subject
	template.TextBody = body.TextBody
	template.HtmlBody = body.HtmlBody

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

	template, err := h.app.templateRepo.Get(split[0], split[1])
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
