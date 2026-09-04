package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/digitalpapyrus/backend/pkg/response"
)

type RegionHandler struct {
	db *sql.DB
}

func NewRegionHandler(db *sql.DB) *RegionHandler {
	return &RegionHandler{db: db}
}

type ProvinceResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type RegencyResponse struct {
	Code         string `json:"code"`
	Name         string `json:"name"`
	ProvinceCode string `json:"province_code"`
}

type DistrictResponse struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	RegencyCode string `json:"regency_code"`
}

type VillageResponse struct {
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	DistrictCode string   `json:"district_code"`
	PostalCodes  []string `json:"postal_codes"`
}

func (h *RegionHandler) GetProvinces(c *gin.Context) {
	rows, err := h.db.Query("SELECT code, name FROM provinces ORDER BY name ASC")
	if err != nil {
		response.InternalError(c, "Failed to fetch provinces: "+err.Error())
		return
	}
	defer rows.Close()

	var list []ProvinceResponse
	for rows.Next() {
		var p ProvinceResponse
		if err := rows.Scan(&p.Code, &p.Name); err == nil {
			list = append(list, p)
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (h *RegionHandler) GetRegenciesByProvince(c *gin.Context) {
	provCode := c.Param("code")
	rows, err := h.db.Query("SELECT code, name, province_code FROM regencies WHERE province_code = $1 ORDER BY name ASC", provCode)
	if err != nil {
		response.InternalError(c, "Failed to fetch regencies: "+err.Error())
		return
	}
	defer rows.Close()

	var list []RegencyResponse
	for rows.Next() {
		var r RegencyResponse
		if err := rows.Scan(&r.Code, &r.Name, &r.ProvinceCode); err == nil {
			list = append(list, r)
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (h *RegionHandler) GetDistrictsByRegency(c *gin.Context) {
	regencyCode := c.Param("code")
	rows, err := h.db.Query("SELECT code, name, regency_code FROM districts WHERE regency_code = $1 ORDER BY name ASC", regencyCode)
	if err != nil {
		response.InternalError(c, "Failed to fetch districts: "+err.Error())
		return
	}
	defer rows.Close()

	var list []DistrictResponse
	for rows.Next() {
		var d DistrictResponse
		if err := rows.Scan(&d.Code, &d.Name, &d.RegencyCode); err == nil {
			list = append(list, d)
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}

func (h *RegionHandler) GetVillagesByDistrict(c *gin.Context) {
	distCode := c.Param("code")
	rows, err := h.db.Query("SELECT code, name, district_code, postal_code FROM villages WHERE district_code = $1 ORDER BY name ASC", distCode)
	if err != nil {
		response.InternalError(c, "Failed to fetch villages: "+err.Error())
		return
	}
	defer rows.Close()

	var list []VillageResponse
	for rows.Next() {
		var v VillageResponse
		var postalCode sql.NullString
		if err := rows.Scan(&v.Code, &v.Name, &v.DistrictCode, &postalCode); err == nil {
			v.PostalCodes = []string{}
			if postalCode.Valid && postalCode.String != "" {
				v.PostalCodes = append(v.PostalCodes, postalCode.String)
			}
			list = append(list, v)
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": list})
}
