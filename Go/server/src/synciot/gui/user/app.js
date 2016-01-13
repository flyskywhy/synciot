// Copyright (C) 2015 Synciot


/*jslint browser: true, continue: true, plusplus: true */
/*global $: false, angular: false, console: false, validLangs: false */

var user = angular.module('user', [
    'pascalprecht.translate',

    'user.core'
]);

var urlbase = 'rest';

user.config(function ($translateProvider) {
    // language and localisation

    $translateProvider.useStaticFilesLoader({
        prefix: 'assets/lang/lang-',
        suffix: '.json'
    });

    $translateProvider.use("zh-CN");
});

function clientCompare(a, b) {
    if (a.id < b.id) {
        return -1;
    }
    return a.id > b.id;
}

function clientMap(l) {
    var m = {};
    l.forEach(function (r) {
        m[r.id] = r;
    });
    return m;
}

function clientList(m) {
    var l = [];
    for (var id in m) {
        l.push(m[id]);
    }
    l.sort(clientCompare);
    return l;
}

function isEmptyObject(obj) {
    var name;
    for (name in obj) {
        return false;
    }
    return true;
}

function debounce(func, wait) {
    var timeout, args, context, timestamp, result, again;

    var later = function () {
        var last = Date.now() - timestamp;
        if (last < wait) {
            timeout = setTimeout(later, wait - last);
        } else {
            timeout = null;
            if (again) {
                again = false;
                result = func.apply(context, args);
                context = args = null;
            }
        }
    };

    return function () {
        context = this;
        args = arguments;
        timestamp = Date.now();
        var callNow = !timeout;
        if (!timeout) {
            timeout = setTimeout(later, wait);
            result = func.apply(context, args);
            context = args = null;
        } else {
            again = true;
        }

        return result;
    };
}
