"use strict";

var annotateApp = angular.module('annotateApp', [
	'ngRoute',
	'annotateControllers',
	'mgcrea.ngStrap',
]);

var timeFormat = 'YYYY-MM-DDTHH:mm:ssZ';


class Annotation {
	constructor(a) {
		a = a || {};
		this.Id = a.Id || "";
		this.Message = a.Message || "";
		this.StartDate = a.StartDate || "";
		this.EndDate = a.EndDate || "";
		this.CreationUser = a.CreationUser;
		this.Url = a.Url || "";
		this.Source = a.Source || "annotate-ui";
		this.Host = a.Host || "";
		this.Owner = a.Owner || "";
		this.Category = a.Category || "";
	}
	setTime() {
		var now = moment().format(timeFormat)
		this.StartDate = now;
		this.EndDate = now;
	}
}



// Reference Struct
	// type Annotation struct {
	// 	Id           string
	// 	Message      string
	// 	StartDate    time.Time
	// 	EndDate      time.Time
	// 	CreationUser string
	// 	Url          *url.URL `json:",omitempty"`
	// 	Source       string
	// 	Host         string
	// 	Owner        string
	// 	Category     string
	// }


annotateApp.config(['$routeProvider', '$locationProvider', '$httpProvider', function($routeProvider, $locationProvider, $httpProvider) {
	$locationProvider.html5Mode(true);
	$routeProvider.
		when('/', {
			title: 'Create',
			templateUrl: 'static/partials/create.html',
			controller: 'CreateCtrl',
		}).
		when('/list', {
			title: 'List',
			templateUrl: 'static/partials/list.html',
			controller: 'ListCtrl',
		}).
		otherwise({
			redirectTo: '/',
		});
}]);

annotateApp.run(['$location', '$rootScope', function($location, $rootScope) {
	// $rootScope.$on('$routeChangeSuccess', function(event, current, previous) {
	// 	$rootScope.title = current.$$route.title;
	// });
}]);

var annotateControllers = angular.module('annotateControllers', [])

annotateControllers.controller('AnnotateCtrl', ['$scope', '$route', '$http', '$rootScope', function($scope, $route, $http, $rootScope) {
	$scope.active = (v) => {
		if (!$route.current) {
			return null;
		}
		if ($route.current.loadedTemplateUrl == 'partials/' + v + '.html') {
			return { active: true };
		}
		return null;
	};
}]);

annotateControllers.controller('CreateCtrl', ['$scope', '$http', '$routeParams', function($scope, $http, $p) {
	if ($p.guid) {
		$http.get('/annotation/' + $p.guid)
			.success((data) => {
				$scope.annotation = new Annotation(data);
			})
			.error((error) => {
				$scope.error = error;
			})
	} else {
		var a = new Annotation();
		a.setTime();
		$scope.annotation = a;
	}
	$scope.submit = () => {
		$http.post('/annotation', $scope.annotation)
			.success((data) => {
				$scope.annotation = new Annotation(data);
				console.log($scope.annotation.Id);
			})
			.error((error) => {
				$scope.error = error;
			})
	};
}]);

annotateControllers.controller('ListCtrl', ['$scope', '$http', function($scope, $http) {
		var EndDate = moment().format(timeFormat)
		var StartDate = moment().subtract(1, "hours").format(timeFormat)
		var params = "StartDate=" + encodeURIComponent(StartDate) + "&EndDate=" + encodeURIComponent(EndDate);
		$http.get('/annotation/query?' + params)
			.success(function(data) {
				$scope.annotations = data;
			})
			.error(function(error) {
				$scope.status = 'Unable to fetch annotations: ' + error;
			});
}]);