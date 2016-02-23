## API

var urlbase = 'rest';

### Server

### Client

#### GET /rest/client/config

    $http.get(urlbase + '/client/config?server=' + encodeURIComponent($scope.thisServerId())).success(function (data) {

这里返回的 data 的数据结构是 UserConfiguration，其中的 id 指的是 client 的数字序号， name 指的是 client 的别名。目前 name 与 id 相同，后续会考虑支持给 client 取别名。

    type ClientConfiguration struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }

    type UserConfiguration struct {
        Clients []ClientConfiguration `json:"clients"`
    }

#### GET /rest/client/status

    $http.get(urlbase + '/client/status?serverId=' + encodeURIComponent($scope.thisServerId())
                                    + ';clientId=' + encodeURIComponent(client)
                                    + ';userIdNum=' + encodeURIComponent($scope.userIdNum)).success(function (data) {

这里输入的 userIdNum 参数是对应数字的字符串 string 变量。

    res["state"] = 表明client的状态是"syncing"或"idle"
    res["out"]  = 整形 Int 变量。表明client回应的历史总次数，也就是保存在最终结果目录“${synciot}/io/out/${Client}/”中的目录数。

#### POST /rest/client/start

    $http.post(urlbase + '/client/start?serverId=' + encodeURIComponent($scope.thisServerId())
                                    + ';userIdNum=' + encodeURIComponent($scope.userIdNum)).success(function () {

这里输入的 userIdNum 参数是用户数字序号的字符串 string 变量。

#### POST /rest/client/stop

    $http.post(urlbase + '/client/stop?serverId=' + encodeURIComponent($scope.thisServerId())
                                   + ';userIdNum=' + encodeURIComponent($scope.userIdNum)).success(function () {

这里输入的 userIdNum 参数是用户数字序号的字符串 string 变量。
