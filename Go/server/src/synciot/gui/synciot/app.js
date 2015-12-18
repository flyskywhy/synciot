// Copyright (C) 2015 Synciot


/*jslint browser: true, continue: true, plusplus: true */
/*global $: false, angular: false, console: false, validLangs: false */

var synciot = angular.module('synciot', [
    'pascalprecht.translate',

    'synciot.core',
    'synciot.folder'
]);

var urlbase = 'rest';

synciot.config(function ($translateProvider) {
    // language and localisation

    $translateProvider.useStaticFilesLoader({
        prefix: 'assets/lang/lang-',
        suffix: '.json'
    });

    $translateProvider.use("zh-CN");
});

function folderCompare(a, b) {
    if (a.id < b.id) {
        return -1;
    }
    return a.id > b.id;
}

function folderMap(l) {
    var m = {};
    l.forEach(function (r) {
        m[r.id] = r;
    });
    return m;
}

function folderList(m) {
    var l = [];
    for (var id in m) {
        l.push(m[id]);
    }
    l.sort(folderCompare);
    return l;
}

function isEmptyObject(obj) {
    var name;
    for (name in obj) {
        return false;
    }
    return true;
}
