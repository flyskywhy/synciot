angular.module('synciot.folder')
    .directive('editFolderModal', function () {
        return {
            restrict: 'A',
            templateUrl: 'synciot/folder/editFolderModalView.html'
        };
});
