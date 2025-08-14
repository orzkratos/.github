package yyle88

import (
	"net/http"
	"os"
	"strings"
	"time"

	restyv2 "github.com/go-resty/resty/v2"
	"github.com/yyle88/erero"
	"github.com/yyle88/neatjson/neatjsons"
	"github.com/yyle88/sortslice"
	"github.com/yyle88/zaplog"
)

type Repo struct {
	Name       string    `json:"name"`
	Link       string    `json:"html_url"`
	Desc       string    `json:"description"`
	Stargazers int       `json:"stargazers_count"`
	PushedAt   time.Time `json:"pushed_at"`
}

func GetGithubRepos(username string) ([]*Repo, error) {
	var repos []*Repo

	// 从环境变量读取 GitHub Token
	githubToken := os.Getenv("GITHUB_TOKEN")

	// 使用 Token 添加 Authorization 请求头
	request := restyv2.New().SetTimeout(time.Minute).R()
	if githubToken != "" {
		request = request.SetHeader("Authorization", "token "+githubToken)
	}
	response, err := request.SetPathParam("username", username).
		SetResult(&repos).
		Get("https://api.github.com/users/{username}/repos")
	if err != nil {
		return nil, erero.Wro(err)
	}
	if response.StatusCode() != http.StatusOK {
		return nil, erero.New(response.Status())
	}
	zaplog.SUG.Debugln(neatjsons.SxB(response.Body()))

	sortslice.SortVStable(repos, func(a, b *Repo) bool {
		if strings.HasPrefix(a.Name, ".") || strings.HasPrefix(b.Name, ".") {
			return !strings.HasPrefix(a.Name, ".")
		} else if a.Name == username || b.Name == username {
			return a.Name != username //当是主页项目时把它排在最后面，避免排的太靠前占据重要的位置
		} else if a.Stargazers > b.Stargazers {
			return true //星多者排前面
		} else if a.Stargazers < b.Stargazers {
			return false //星少者排后面
		} else {
			return a.PushedAt.After(b.PushedAt) //同样星星时最近有更新的排前面
		}
	})

	zaplog.SUG.Debugln(neatjsons.S(repos))
	return repos, nil
}

type Organization struct {
	Name      string `json:"login"`     // 组织名称
	Link      string `json:"url"`       // 组织链接
	ReposLink string `json:"repos_url"` // 组织链接
}

type Membership struct {
	Role  string `json:"role"`  // "admin", "member"
	State string `json:"state"` // "active", "pending"
}

func checkOrganizationOwnership(orgName, username string) (bool, error) {
	// 从环境变量读取 GitHub Token
	githubToken := os.Getenv("GITHUB_TOKEN")

	// 使用 Token 添加 Authorization 请求头
	request := restyv2.New().SetTimeout(time.Minute).R()
	if githubToken != "" {
		request = request.SetHeader("Authorization", "token "+githubToken)
	}

	var membership Membership
	response, err := request.SetPathParams(map[string]string{
		"org":      orgName,
		"username": username,
	}).SetResult(&membership).
		Get("https://api.github.com/orgs/{org}/memberships/{username}")

	if err != nil {
		return false, erero.Wro(err)
	}

	// 如果状态码是200且role是admin，说明用户是owner
	if response.StatusCode() == http.StatusOK {
		zaplog.SUG.Debugf("Organization %s membership for %s: role=%s, state=%s", orgName, username, membership.Role, membership.State)
		return membership.Role == "admin" && membership.State == "active", nil
	}

	// 其他状态码表示不是成员或无权限
	return false, nil
}

func GetOrganizations(username string) ([]*Organization, error) {
	var allOrganizations []*Organization

	// 从环境变量读取 GitHub Token
	githubToken := os.Getenv("GITHUB_TOKEN")

	// 使用 Token 添加 Authorization 请求头
	request := restyv2.New().SetTimeout(time.Minute).R()
	if githubToken != "" {
		request = request.SetHeader("Authorization", "Bearer "+githubToken)
	}

	// 请求获取用户的组织信息
	response, err := request.SetPathParam("username", username).
		SetResult(&allOrganizations).
		Get("https://api.github.com/users/{username}/orgs")
	if err != nil {
		return nil, erero.Wro(err)
	}
	if response.StatusCode() != http.StatusOK {
		return nil, erero.New(response.Status())
	}
	zaplog.SUG.Debugln(neatjsons.SxB(response.Body()))

	// 暂时返回所有组织以测试其他功能
	zaplog.SUG.Debugln(neatjsons.S(allOrganizations))
	return allOrganizations, nil
}

func GetOrganizationRepos(orgName string) ([]*Repo, error) {
	var repos []*Repo

	// 从环境变量读取 GitHub Token
	githubToken := os.Getenv("GITHUB_TOKEN")

	// 使用 Token 添加 Authorization 请求头
	request := restyv2.New().SetTimeout(time.Minute).R()
	if githubToken != "" {
		request = request.SetHeader("Authorization", "token "+githubToken)
	}
	response, err := request.SetPathParam("org", orgName).
		SetResult(&repos).
		Get("https://api.github.com/orgs/{org}/repos")
	if err != nil {
		return nil, erero.Wro(err)
	}
	if response.StatusCode() != http.StatusOK {
		return nil, erero.New(response.Status())
	}
	zaplog.SUG.Debugln(neatjsons.SxB(response.Body()))

	sortslice.SortVStable(repos, func(a, b *Repo) bool {
		if strings.HasPrefix(a.Name, ".") || strings.HasPrefix(b.Name, ".") {
			return !strings.HasPrefix(a.Name, ".")
		} else if a.Name == orgName || b.Name == orgName {
			return a.Name == orgName //当项目名称与组织名称相同时，说明是主项目，因此需要放在前面
		} else if a.Stargazers > b.Stargazers {
			return true //星多者排前面
		} else if a.Stargazers < b.Stargazers {
			return false //星少者排后面
		} else {
			return a.PushedAt.After(b.PushedAt) //同样星星时最近有更新的排前面
		}
	})

	zaplog.SUG.Debugln(neatjsons.S(repos))
	return repos, nil
}
