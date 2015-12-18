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
        $scope.deviceName = "(server)";
        $scope.folders = {};

        $scope.emitHTTPError = function (data, status, headers, config) {
            $scope.$emit('HTTPError', {data: data, status: status, headers: headers, config: config});
        };

        function updateLocalConfig(config) {
            var hasConfig = !isEmptyObject($scope.config);

            $scope.config = config;
            $scope.folders = folderMap($scope.config.folders);

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

        $scope.thisDeviceName = function () {
            return $scope.deviceName;
        };

        $scope.saveConfig = function () {
            var cfg = JSON.stringify($scope.config);
            var opts = {
                headers: {
                    'Content-Type': 'application/json'
                }
            };
            $http.post(urlbase + '/system/config', cfg, opts).success(function () {
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

            $('#editFolder').modal('hide');
            folderCfg = $scope.currentFolder;

            $scope.folders[folderCfg.id] = folderCfg;
            $scope.config.folders = folderList($scope.folders);

            $scope.saveConfig();
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

        // pseudo main. called on all definitions assigned
        initController();
    });
