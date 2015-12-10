// Copyright (C) 2015 Synciot


/*jslint browser: true, continue: true, plusplus: true */
/*global $: false, angular: false, console: false, validLangs: false */

var synciot = angular.module('synciot', [
    'pascalprecht.translate',

    'synciot.core'
]);

synciot.config(function ($translateProvider) {
    // language and localisation

    $translateProvider.useStaticFilesLoader({
        prefix: 'assets/lang/lang-',
        suffix: '.json'
    });

    $translateProvider.use("zh-CN");
});
