angular.module('synciot.core')
    .config(function($locationProvider) {
        $locationProvider.html5Mode(true).hashPrefix('!');
    })
    .controller('SynciotController', function ($scope, $http, $location) {
        'use strict';

        // private/helper definitions

        function initController() {
            $scope.refresh();
            setInterval($scope.refresh, 10000);
            refreshConfig();
        }

        // public/scope definitions

        $scope.config = {};
        $scope.configInSync = true;
        $scope.model = {};
        $scope.pageName = "Administrator";
        $scope.folders = {};
//        $scope.syncthingGuiPorts = {};
//        $scope.syncthingProtocolPorts = {};

        $scope.emitHTTPError = function (data, status, headers, config) {
            $scope.$emit('HTTPError', {data: data, status: status, headers: headers, config: config});
        };

        var debouncedFuncs = {};

        function refreshFolder(folder) {
            var key = "refreshFolder" + folder;
            if (!debouncedFuncs[key]) {
                debouncedFuncs[key] = debounce(function () {
                    $http.get(urlbase + '/stats/folder?folder=' + encodeURIComponent(folder)).success(function (data) {
                        $scope.model[folder] = data;
                        console.log("refreshFolder", folder, data);
                    }).error($scope.emitHTTPError);
                }, 1000, true);
            }
            debouncedFuncs[key]();
        }

        function updateLocalConfig(config) {
            var hasConfig = !isEmptyObject($scope.config);

            $scope.config = config;
            $scope.folders = folderMap($scope.config.folders);
            Object.keys($scope.folders).forEach(function (folder) {
                refreshFolder(folder);
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
            $http.get(urlbase + '/system/config').success(function (data) {
                updateLocalConfig(data);
                console.log("refreshConfig", data);
            }).error($scope.emitHTTPError);
        }

        $scope.refresh = function () {
            refreshSystem();
        };

        $scope.folderStatus = function (folderCfg) {
            if (typeof $scope.model[folderCfg.id] === 'undefined') {
                return 'unknown';
            }

            if (!$scope.model[folderCfg.id].state) {
                return 'unknown';
            }

            var state = '' + $scope.model[folderCfg.id].state;

            return state;
        };

        $scope.syncthingGuiAddress = function (folderCfg) {
            if (typeof $scope.model[folderCfg.id] === 'undefined') {
                return 'unknown';
            }

            if (!$scope.model[folderCfg.id].syncthingGuiPort) {
                return 'unknown';
            }

            var address = $location.protocol() + '://' + $location.host() + ':' + $scope.model[folderCfg.id].syncthingGuiPort;

            return address;
        };

        $scope.userGuiAddress = function (folderCfg) {
            var address = $location.protocol() + '://' + $location.host() + ':' + $location.port() + '/user-' + folderCfg.id + '.html';

            return address;
        };

        $scope.thisPageName = function () {
            return $scope.pageName;
        };

        $scope.saveConfig = function () {
            var cfg = angular.toJson($scope.config);
            var opts = {
                headers: {
                    'Content-Type': 'application/json'
                }
            };
            $http.post(urlbase + '/system/config', cfg, opts).success(function () {
                refreshConfig();
            }).error($scope.emitHTTPError);
        };

        $scope.folderList = function () {
            return folderList($scope.folders);
        };

        $scope.directoryList = ['~/synciot', 'D:\\synciot'];

        $scope.editFolder = function (folderCfg) {
            $scope.currentFolder = angular.copy(folderCfg);
            if ($scope.currentFolder.path.slice(-1) == $scope.system.pathSeparator) {
                $scope.currentFolder.path = $scope.currentFolder.path.slice(0, -1);
            }

            $scope.editingExisting = true;
            $scope.folderEditor.$setPristine();
            $('#editFolder').modal();
        };

        $scope.addFolder = function () {
            $scope.currentFolder = {
            };
            $scope.editingExisting = false;
            $scope.folderEditor.$setPristine();
            $('#editFolder').modal();
        };

        $scope.saveFolder = function () {
            var folderCfg;

            if ($scope.currentFolder.path.trim().charAt(0) == '~') {
                $scope.currentFolder.path = $scope.system.tilde + $scope.currentFolder.path.trim().substring(1);
            }

            $('#editFolder').modal('hide');
            folderCfg = $scope.currentFolder;

            $scope.folders[folderCfg.id] = folderCfg;
            $scope.config.folders = folderList($scope.folders);

            $http.post(urlbase + '/system/generate?path=' + encodeURIComponent(folderCfg.path)
                                                 + ';id=' + encodeURIComponent(folderCfg.id)).success(function () {
                $scope.saveConfig();
            }).error($scope.emitHTTPError);

        };

        $scope.deleteFolder = function (id) {
            $('#editFolder').modal('hide');
            if (!$scope.editingExisting) {
                return;
            }

            delete $scope.folders[id];
            $scope.config.folders = folderList($scope.folders);

            $scope.saveConfig();
        };

        $scope.about = function () {
            $('#about').modal('show');
        };

//            for (var id in $scope.folders) {
//                $scope.folders[id].devices = $scope.folders[id].devices.filter(function (n) {
//                    return n.deviceID !== $scope.currentDevice.deviceID;
//                });
//            }



        $scope.stopSyncthing = function (folderCfg) {
            $http.post(urlbase + "/system/stop?folder=" + encodeURIComponent(folderCfg.id)).success(function () {
                $scope.model[folderCfg.id].state = 'stopped';
            }).error($scope.emitHTTPError);
        };

        $scope.startSyncthing = function (folderCfg) {
            $http.post(urlbase + '/system/start?folder=' + encodeURIComponent(folderCfg.id)).success(function () {
                $scope.model[folderCfg.id].state = 'running';
            }).error($scope.emitHTTPError);
        };

        // pseudo main. called on all definitions assigned
        initController();
    });
