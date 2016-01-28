angular.module('user.core')
    .config(function($locationProvider) {
        $locationProvider.html5Mode(true).hashPrefix('!');
    })
    .controller('UserController', function ($scope, $http, $location) {
        'use strict';

        // private/helper definitions

        function initController() {
            $scope.refresh();
            setInterval($scope.refresh, 10000);
            refreshConfig();
        }

        // public/scope definitions

        $scope.config = {};
        $scope.model = {};
        $scope.pageName = "User";
        $scope.clients = {};
        $scope.clientList = [];
        $scope.checkboxMasterDisplay = true;
        $scope.checkboxMasterLogical = true;
        $scope.startStopWaitNextRefreshClient = false;

        $scope.emitHTTPError = function (data, status, headers, config) {
            $scope.$emit('HTTPError', {data: data, status: status, headers: headers, config: config});
        };

        var debouncedFuncs = {};

        function refreshClient(client) {
            var key = "refreshClient" + client;
            if (!debouncedFuncs[key]) {
                debouncedFuncs[key] = debounce(function () {
                    $http.get(urlbase + '/client/status?serverId=' + encodeURIComponent($scope.thisServerId())
                                                    + ';clientId=' + encodeURIComponent(client)).success(function (data) {
                        $scope.model[client] = data;
                        $scope.startStopWaitNextRefreshClient = false;
                        console.log("refreshClient", client, data);
                    }).error($scope.emitHTTPError);
                }, 1000, true);
            }
            debouncedFuncs[key]();
        }

        function updateLocalConfig(config) {
            var hasConfig = !isEmptyObject($scope.config);

            $scope.config = config;
            $scope.clients = clientMap($scope.config.clients);
            $scope.clientList = clientList($scope.clients);
            $scope.clientList.forEach(function (client) {
                client.checkboxSlaveDisplay = true;
                client.checkboxSlaveLogical = true;
            });
            Object.keys($scope.clients).forEach(function (client) {
                refreshClient(client);
            });

            if (!hasConfig) {
                $scope.$emit('ConfigLoaded');
            }
        }

        function refreshSystem() {
            $http.get(urlbase + '/system/status').success(function (data) {
                $scope.system = data;

                console.log("refreshSystem", data);
            }).error($scope.emitHTTPError);
        }

        function refreshConfig() {
            $http.get(urlbase + '/client/config?server=' + encodeURIComponent($scope.thisServerId())).success(function (data) {
                updateLocalConfig(data);
                console.log("refreshConfig", data);
            }).error($scope.emitHTTPError);
        }

        $scope.refresh = function () {
            refreshSystem();

            Object.keys($scope.clients).forEach(function (client) {
                refreshClient(client);
            });
        };

        $scope.clientStatus = function (client) {
            var state = 'unknown'

            if (typeof $scope.model[client.id] === 'undefined') {
                return 'unknown';
            }

            if (!$scope.model[client.id].state) {
                return 'unknown';
            }

            state = '' + $scope.model[client.id].state;
            return state;
        };

        $scope.clientsStatus = function () {
            var state = 'unknown'

            for (var i in $scope.clientList) {
                var client = $scope.clientList[i];
                if (client.checkboxSlaveLogical == true) {
                    if (typeof $scope.model[client.id] === 'undefined') {
                        return 'unknown';
                    }

                    if (!$scope.model[client.id].state) {
                        return 'unknown';
                    }

                    if ($scope.startStopWaitNextRefreshClient) {
                        return 'unknown';
                    }

                    state = '' + $scope.model[client.id].state;
                    if (state === 'running') {
                        return 'running'
                    }
                }
            }

            return state;
        };

        $scope.administratorGuiAddress = function () {
            return $location.protocol() + '://' + $location.host() + ':' + $location.port();
        };

        $scope.thisServerId = function () {
            var path = $location.path()
            return path.substr(6, path.length-11);
        };

        $scope.thisPageName = function () {
            return $scope.pageName;
        };

        $scope.about = function () {
            $('#about').modal('show');
        };

        $scope.stopClient = function () {
            if ($scope.checkboxMasterLogical == true) {
                $http.post(urlbase + '/client/stop?serverId=' + encodeURIComponent($scope.thisServerId())).success(function () {
                    $scope.startStopWaitNextRefreshClient = true;
                }).error($scope.emitHTTPError);
            } else {
                var clientIds = [];

                for (var i in $scope.clientList) {
                    var client = $scope.clientList[i];
                    if (client.checkboxSlaveLogical == true) {
                        clientIds.push(client.id);
                    }
                }

                var opts = {
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };
                $http.post(urlbase + '/client/stop?serverId=' + encodeURIComponent($scope.thisServerId()), angular.toJson(clientIds), opts).success(function () {
                    $scope.startStopWaitNextRefreshClient = true;
                }).error($scope.emitHTTPError);
            }
        };

        $scope.startClient = function () {
            if ($scope.checkboxMasterLogical == true) {
                $http.post(urlbase + '/client/start?serverId=' + encodeURIComponent($scope.thisServerId())).success(function () {
                    $scope.startStopWaitNextRefreshClient = true;
                }).error($scope.emitHTTPError);
            } else {
                var clientIds = [];

                for (var i in $scope.clientList) {
                    var client = $scope.clientList[i];
                    if (client.checkboxSlaveLogical == true) {
                        clientIds.push(client.id);
                    }
                }

                var opts = {
                    headers: {
                        'Content-Type': 'application/json'
                    }
                };
                $http.post(urlbase + '/client/start?serverId=' + encodeURIComponent($scope.thisServerId()), angular.toJson(clientIds), opts).success(function () {
                    $scope.startStopWaitNextRefreshClient = true;
                }).error($scope.emitHTTPError);
            }
        };

        $scope.checkboxAll = function () {
            if ($scope.checkboxMasterLogical == true) {
                $scope.checkboxMasterLogical = false;
                $scope.clientList.forEach(function (client) {
                    client.checkboxSlaveDisplay = false;
                    client.checkboxSlaveLogical = false;
                });
            } else {
                $scope.checkboxMasterLogical = true;
                $scope.clientList.forEach(function (client) {
                    client.checkboxSlaveDisplay = true;
                    client.checkboxSlaveLogical = true;
                });
            }
        };

        $scope.checkboxOne = function (client) {
            $scope.checkboxMasterDisplay = false
            $scope.checkboxMasterLogical = false;
            if (client.checkboxSlaveLogical == true) {
                client.checkboxSlaveLogical = false;
            } else {
                client.checkboxSlaveLogical = true;
            }
        };

        // pseudo main. called on all definitions assigned
        initController();
    });
