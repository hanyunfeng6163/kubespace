package serializers

type UserCreateSerializers struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type UserSerializers struct {
	UserName string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

type ClusterCreateSerializers struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type ClusterUpdateSerializers struct {
	KubeConfig string `json:"kubeconfig"`
}

type DeleteClusterSerializers struct {
	Id string `json:"name"`
}

type DeleteUserSerializers struct {
	Name string `json:"name"`
}

type DeleteRoleSerializers struct {
	Name string `json:"name"`
}

type ApplyYamlSerializers struct {
	YamlStr string `json:"yaml"`
}

type UserRoleSerializers struct {
	UserId  uint   `json:"user_id" form:"user_id"`
	Scope   string `json:"scope" form:"scope"`
	ScopeId uint   `json:"scope_id" form:"scope_id"`
	Role    string `json:"role" form:"from"`
}

type UserRoleUpdateSerializers struct {
	UserIds []uint `json:"user_ids" form:"user_ids"`
	*UserRoleSubScope
	SubScopes []*UserRoleSubScope `json:"sub_scopes" form:"sub_scopes"`
}

type UserRoleSubScope struct {
	Scope      string `json:"scope" form:"scope"`
	ScopeId    uint   `json:"scope_id" form:"scope_id"`
	ScopeRegex string `json:"scope_regex" form:"scope_regex"`
	Role       string `json:"role" form:"role"`
}

type UpdatePasswordSerializers struct {
	OriginPassword string `json:"origin_password" form:"origin_password"`
	NewPassword    string `json:"new_password" form:"new_password"`
}
