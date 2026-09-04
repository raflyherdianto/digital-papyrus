package handler

import (
	"net/http"

	"github.com/digitalpapyrus/backend/internal/repository"
	"github.com/digitalpapyrus/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	repo *repository.SettingRepo
}

func NewSettingHandler(repo *repository.SettingRepo) *SettingHandler {
	return &SettingHandler{repo: repo}
}

// GetSettings fetches the public settings
func (h *SettingHandler) GetSettings(c *gin.Context) {
	settings, err := h.repo.GetAllSettings()
	if err != nil {
		response.InternalError(c, "Failed to retrieve settings")
		return
	}

	// For compatibility with frontend interface Settings which expects "tax" and "discount"
	if val, ok := settings["tax_percentage"]; ok {
		settings["tax"] = val
	}
	if val, ok := settings["discount_percentage"]; ok {
		settings["discount"] = val
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

type UpdateSettingsRequest struct {
	OriginVillageCode  string `json:"origin_village_code"`
	TaxPercentage      string `json:"tax_percentage"`
	Tax                string `json:"tax"`
	ServiceFee         string `json:"service_fee"`
	DiscountPercentage string `json:"discount_percentage"`
	Discount           string `json:"discount"`
	OriginPhone        string `json:"origin_phone"`
	OriginAddress      string `json:"origin_address"`
	OriginProvince     string `json:"origin_province"`
	OriginCity         string `json:"origin_city"`
	OriginDistrict     string `json:"origin_district"`
	OriginZipCode      string `json:"origin_zip_code"`
	BankName           string `json:"bank_name"`
	BankAccountNumber  string `json:"bank_account_number"`
	BankAccountHolder  string `json:"bank_account_holder"`
}

// UpdateSettings allows superadmins to update settings
func (h *SettingHandler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid input data", nil)
		return
	}

	taxVal := req.TaxPercentage
	if taxVal == "" {
		taxVal = req.Tax
	}
	discountVal := req.DiscountPercentage
	if discountVal == "" {
		discountVal = req.Discount
	}

	settings := map[string]string{
		"origin_village_code":  req.OriginVillageCode,
		"tax_percentage":       taxVal,
		"service_fee":          req.ServiceFee,
		"discount_percentage":  discountVal,
		"origin_phone":         req.OriginPhone,
		"origin_address":       req.OriginAddress,
		"origin_province":      req.OriginProvince,
		"origin_city":          req.OriginCity,
		"origin_district":      req.OriginDistrict,
		"origin_zip_code":      req.OriginZipCode,
		"bank_name":            req.BankName,
		"bank_account_number":  req.BankAccountNumber,
		"bank_account_holder":  req.BankAccountHolder,
	}

	if err := h.repo.UpdateSettings(settings); err != nil {
		response.InternalError(c, "Failed to update settings")
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Settings updated successfully"})
}
