angular.module('synciot.core')
    .directive('modal', function () {
        return {
            restrict: 'E',
            templateUrl: 'modal.html',
            replace: true,
            transclude: true,
            scope: {
                title: '@',
                status: '@',
                icon: '@',
                close: '@',
                large: '@'
            }
        };
    });
