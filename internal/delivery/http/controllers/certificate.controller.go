package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CertificateController struct {
	svc *service.CertificateService
}

func NewCertificateController(svc *service.CertificateService) *CertificateController {
	return &CertificateController{svc: svc}
}

func (c *CertificateController) CreateCertificate(ctx *gin.Context) {
	idParam := ctx.Param("id")
	personnelID, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid personnel id format")
		return
	}

	var req service.CreateCertificateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	req.PersonnelID = personnelID

	cert, err := c.svc.Create(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, http.StatusCreated, "certificate created successfully", cert)
}

func (c *CertificateController) ListCertificates(ctx *gin.Context) {
	idParam := ctx.Param("id")
	personnelID, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid personnel id format")
		return
	}

	certs, err := c.svc.FindByPersonnelID(ctx.Request.Context(), personnelID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get certificates")
		return
	}

	response.Success(ctx, http.StatusOK, "certificates retrieved successfully", certs)
}

func (c *CertificateController) UpdateCertificate(ctx *gin.Context) {
	certIdParam := ctx.Param("certId")
	certID, err := bson.ObjectIDFromHex(certIdParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid certificate id format")
		return
	}

	var req service.CreateCertificateInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	cert, err := c.svc.Update(ctx.Request.Context(), certID, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, http.StatusOK, "certificate updated successfully", cert)
}

func (c *CertificateController) DeleteCertificate(ctx *gin.Context) {
	certIdParam := ctx.Param("certId")
	certID, err := bson.ObjectIDFromHex(certIdParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid certificate id format")
		return
	}

	err = c.svc.Delete(ctx.Request.Context(), certID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to delete certificate")
		return
	}

	response.Success(ctx, http.StatusOK, "certificate deleted successfully", nil)
}
