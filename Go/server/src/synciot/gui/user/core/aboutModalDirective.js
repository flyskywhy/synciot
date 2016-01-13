angular.module('user.core')
    .directive('aboutModal', function () {
        return {
            restrict: 'A',
            templateUrl: 'user/core/aboutModalView.html'
        };
});
