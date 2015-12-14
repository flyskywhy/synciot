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

        $scope.thisDeviceName = function () {
            return $scope.deviceName;
        };

        $scope.about = function () {
            $('#about').modal('show');
        };

        // pseudo main. called on all definitions assigned
        initController();
    });
