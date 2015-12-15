angular.module('synciot.core')
    .config(function($locationProvider) {
        $locationProvider.html5Mode(true).hashPrefix('!');
    })
    .controller('SynciotController', function ($scope, $http, $location) {
        'use strict';

        // private/helper definitions

        function initController() {
            setInterval($scope.refresh, 10000);
        }

        // public/scope definitions

        $scope.configInSync = true;
        $scope.deviceName = "(server)";
        $scope.folders = {};

        $scope.emitHTTPError = function (data, status, headers, config) {
            $scope.$emit('HTTPError', {data: data, status: status, headers: headers, config: config});
        };

        function refreshSystem() {
            $http.get(urlbase + '/system/status').success(function (data) {
                $scope.system = data;

                console.log("refreshSystem", data);
            }).error($scope.emitHTTPError);
        }

        $scope.refresh = function () {
            refreshSystem();
        };

        $scope.thisDeviceName = function () {
            return $scope.deviceName;
        };

        $scope.folderList = function () {
            return folderList($scope.folders);
        };

        $scope.directoryList = ['~/synciot', 'D:\\synciot'];

        $scope.addFolder = function () {
            $scope.currentFolder = {
            };
            $scope.editingExisting = false;
            $scope.folderEditor.$setPristine();
            $('#editFolder').modal();
        };

        $scope.about = function () {
            $('#about').modal('show');
        };

        // pseudo main. called on all definitions assigned
        initController();
    });
