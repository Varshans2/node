package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/mysteriumnetwork/node/tequilapi/contract"
	"github.com/mysteriumnetwork/node/ui"
	"net/http"
)

type NodeUIEndpoints struct {
	versionManager *ui.VersionManager
}

// TODO good time to introduce a common error response

func NewNodeUIEndpoints(versionManager *ui.VersionManager) *NodeUIEndpoints {
	return &NodeUIEndpoints{
		versionManager: versionManager,
	}
}

func (n *NodeUIEndpoints) remoteVersions(c *gin.Context) {
	remote, err := n.versionManager.ListRemote()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, remote)
}

func (n *NodeUIEndpoints) localVersions(c *gin.Context) {
	local, err := n.versionManager.ListLocal()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, local)
}

func (n *NodeUIEndpoints) switchVersion(c *gin.Context) {
	var req contract.SwitchNodeUIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "could not parse request",
		})
		return
	}

	if err := req.Valid(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}

	if err := n.versionManager.SwitchTo(req.Version); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}

	c.AbortWithStatus(200)
}
func (n *NodeUIEndpoints) download(c *gin.Context) {
	var req contract.DownloadNodeUIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
			"errorMessage": "could not parse request",
		})
		return
	}

	if err := req.Valid(); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}

	if err := n.versionManager.Download(req.Version); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]interface{}{
			"errorMessage": err.Error(),
		})
		return
	}

	c.AbortWithStatus(200)
}

// AddRoutesForNodeUI provides controls for nodeUI management via tequilapi
func AddRoutesForNodeUI(versionManager *ui.VersionManager) func(*gin.Engine) error {
	endpoints := NewNodeUIEndpoints(versionManager)

	return func(e *gin.Engine) error {
		v1Group := e.Group("/ui")
		{
			v1Group.GET("/local-versions", endpoints.localVersions)
			v1Group.GET("/remote-versions", endpoints.remoteVersions)
			v1Group.POST("/switch-version", endpoints.switchVersion)
			v1Group.POST("/download-version", endpoints.download)
		}
		return nil
	}
}
