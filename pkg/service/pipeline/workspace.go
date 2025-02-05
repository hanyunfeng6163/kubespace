package pipeline

import (
	"context"
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/pipeline/schemas"
	"github.com/kubespace/kubespace/pkg/third/git"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"regexp"
	"time"
)

type WorkspaceService struct {
	models *model.Models
}

func NewWorkspaceService(models *model.Models) *WorkspaceService {
	return &WorkspaceService{
		models: models,
	}
}

func (w *WorkspaceService) getCodeName(codeType string, codeUrl string) string {
	var re *regexp.Regexp
	if codeType == types.WorkspaceCodeTypeGit {
		re, _ = regexp.Compile("git@[\\w\\.]+:/?([\\w/\\-_]+)[\\.git]*")
	} else {
		re, _ = regexp.Compile("http[s]?://[\\w\\.:]+/([\\w/\\-_]+)[.git]*")
	}
	codeName := re.FindStringSubmatch(codeUrl)
	if len(codeName) < 2 {
		return ""
	}
	return codeName[1]
}

func (w *WorkspaceService) checkCodeUrl(codeType string, codeUrl string) bool {
	var re *regexp.Regexp
	if codeType == types.WorkspaceCodeTypeGit {
		re, _ = regexp.Compile("git@[\\w\\.]+:/?([\\w/\\-_]+)[\\.git]*")
	} else {
		re, _ = regexp.Compile("http[s]?://[\\w\\.:]+/([\\w/\\-_]+)[.git]*")
	}
	return re.MatchString(codeUrl)
}

func (w *WorkspaceService) defaultCodePipelines() ([]*types.Pipeline, error) {
	branchPipeline := &types.Pipeline{
		Name:       "分支流水线",
		CreateUser: "admin",
		UpdateUser: "admin",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Sources: types.PipelineSources{
			&types.PipelineSource{
				Type:       types.WorkspaceTypeCode,
				BranchType: types.PipelineBranchTypeBranch,
				Operator:   types.PipelineTriggerOperatorExclude,
				Branch:     "master",
			},
		},
		Stages: []*types.PipelineStage{
			{
				Name:        "构建代码镜像",
				TriggerMode: types.StageTriggerModeAuto,
				Jobs: types.PipelineJobs{
					&types.PipelineJob{
						Name:      "构建代码镜像",
						PluginKey: types.BuiltinPluginBuildCodeToImage,
						Params:    map[string]interface{}{},
					},
				},
			},
		},
	}
	masterPipeline := &types.Pipeline{
		Name:       "主干流水线",
		CreateUser: "admin",
		UpdateUser: "admin",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
		Sources: types.PipelineSources{
			&types.PipelineSource{
				Type:       types.WorkspaceTypeCode,
				BranchType: types.PipelineBranchTypeBranch,
				Operator:   types.PipelineTriggerOperatorEqual,
				Branch:     "master",
			},
		},
		Stages: []*types.PipelineStage{
			{
				Name:        "构建代码镜像",
				TriggerMode: types.StageTriggerModeAuto,
				Jobs: types.PipelineJobs{
					&types.PipelineJob{
						Name:      "构建代码镜像",
						PluginKey: types.BuiltinPluginBuildCodeToImage,
						Params:    map[string]interface{}{},
					},
				},
			},
			{
				Name:        "发布",
				TriggerMode: types.StageTriggerModeManual,
				Jobs: types.PipelineJobs{
					&types.PipelineJob{
						Name:      "发布",
						PluginKey: types.BuiltinPluginRelease,
						Params:    map[string]interface{}{},
					},
				},
			},
		},
	}
	return []*types.Pipeline{branchPipeline, masterPipeline}, nil
}

func (w *WorkspaceService) Create(workspaceSer *serializers.WorkspaceSerializer, user *types.User) *utils.Response {
	var err error
	if workspaceSer.Type == types.WorkspaceTypeCode {
		if !w.checkCodeUrl(workspaceSer.CodeType, workspaceSer.CodeUrl) {
			return &utils.Response{Code: code.ParamsError, Msg: "代码地址格式不正确"}
		}
		workspaceSer.Name = w.getCodeName(workspaceSer.CodeType, workspaceSer.CodeUrl)
		secret, err := w.models.SettingsSecretManager.Get(workspaceSer.CodeSecretId)
		if err != nil {
			return &utils.Response{Code: code.GetError, Msg: err.Error()}
		}
		gitcli, err := git.NewClient(workspaceSer.CodeType, workspaceSer.ApiUrl, &types.Secret{
			Type:        secret.Type,
			User:        secret.User,
			Password:    secret.Password,
			PrivateKey:  secret.PrivateKey,
			AccessToken: secret.AccessToken,
		})
		if err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
		}
		// 获取代码仓库分支，验证是否可以连通
		if _, err = gitcli.ListRepoBranches(context.Background(), workspaceSer.CodeUrl); err != nil {
			return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
		}
	}
	if workspaceSer.Name == "" {
		return &utils.Response{Code: code.ParamsError, Msg: "解析代码地址失败，未获取到代码库名称"}
	}
	workspace := &types.PipelineWorkspace{
		Name:        workspaceSer.Name,
		Description: workspaceSer.Description,
		Type:        workspaceSer.Type,
		Code: &types.PipelineWorkspaceCode{
			Type:     workspaceSer.CodeType,
			ApiUrl:   workspaceSer.ApiUrl,
			CloneUrl: workspaceSer.CodeUrl,
			SecretId: workspaceSer.CodeSecretId,
		},
		CreateUser: user.Name,
		UpdateUser: user.Name,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	resp := &utils.Response{Code: code.Success}
	var defaultPipeline []*types.Pipeline
	if workspace.Type == types.WorkspaceTypeCode {
		defaultPipeline, err = w.defaultCodePipelines()
		if err != nil {
			return &utils.Response{Code: code.CreateError, Msg: "创建默认流水线失败: " + err.Error()}
		}
	}
	workspace, err = w.models.PipelineWorkspaceManager.Create(workspace, defaultPipeline)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	resp.Data = workspace
	return resp
}

func (w *WorkspaceService) ListGitRepos(params *schemas.ListGitReposParams) *utils.Response {
	secret, err := w.models.SettingsSecretManager.Get(params.SecretId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	gitcli, err := git.NewClient(params.GitType, params.ApiUrl, &types.Secret{
		Type:        secret.Type,
		User:        secret.User,
		Password:    secret.Password,
		PrivateKey:  secret.PrivateKey,
		AccessToken: secret.AccessToken,
	})
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	repos, err := gitcli.ListRepositories(context.Background())
	if err != nil {
		return &utils.Response{Code: code.RequestError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: repos}
}
