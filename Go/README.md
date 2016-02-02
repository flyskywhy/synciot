## API

var urlbase = 'rest';

### Server

### Client

#### GET /rest/client/config

    $http.get(urlbase + '/client/config?server=' + encodeURIComponent($scope.thisServerId())).success(function (data) {

这里返回的 data 的数据结构是 UserConfiguration，其中的 id 和 name 对应的是 Syncthing 中的 Device 的 id 和 name

    type ClientConfiguration struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }

    type UserConfiguration struct {
        Clients []ClientConfiguration `json:"clients"`
    }

#### GET /rest/client/status

    $http.get(urlbase + '/client/status?serverId=' + encodeURIComponent($scope.thisServerId())
                                                           + ';clientId=' + encodeURIComponent(client)).success(function (data) {

这里返回的 data 的数据结构是 UserConfiguration，其中的 id 和 name 对应的是 Syncthing 中的 Device 的 id 和 name

    res["state"] = 表明client的状态是"syncing"或"idle"
    res["out"]  = 表明client回应的历史总次数，也就是保存在最终结果目录“${synciot}/io/out/${Client}-temp/”中的目录数。

#### POST /rest/client/start

    $http.post(urlbase + '/client/start?serverId=' + encodeURIComponent($scope.thisServerId())).success(function () {

#### POST /rest/client/stop

    $http.post(urlbase + '/client/stop?serverId=' + encodeURIComponent($scope.thisServerId())).success(function () {
