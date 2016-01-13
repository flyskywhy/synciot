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
        }

        // public/scope definitions

        $scope.pageName = "User";

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

        $scope.stopClient = function (clients) {
        };

        $scope.startClient = function (clients) {
        };

        // pseudo main. called on all definitions assigned
        initController();
    });
