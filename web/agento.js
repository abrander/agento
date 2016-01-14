var agento = angular.module('agento', ['ngResource', 'ui.bootstrap']);

var Agento = {};

/**
 * Object to hold all angular controllers
 */
Agento.Controller = {};

/**
 * Object to hold all angular directives
 */
Agento.Directive = {};

/**
 * Object to hold all angular factories
 */
Agento.Factory = {};

/**
 * Object to hold all angular filters
 */
Agento.Filter = {};

/**
 * Object to hold all angular services
 */
Agento.Service = {};

/**
 * @constructor
 * @param {*} $resource
 * @ngInject
 * @suppress {checkTypes}
 */
Agento.Factory.HostService = function($resource) {
	return $resource('/api/host/:id', {id: '@id'}, {
		save: {
			url: '/api/host/new',
			method: 'POST'
		}
	});
};

agento.factory('HostService', Agento.Factory.HostService);

/**
 * @constructor
 * @param {*} $resource
 * @ngInject
 * @suppress {checkTypes}
 */
Agento.Factory.MonitorService = function($resource) {
	return $resource('/api/monitor/:id', {id: '@id'}, {
		save: {
			url: '/api/monitor/new',
			method: 'POST'
		}
	});
};

agento.factory('MonitorService', Agento.Factory.MonitorService);

/**
 * @ngInject
 * @constructor
 */
Agento.Controller.MainController = function(HostService, MonitorService, $http, $uibModal, $scope) {
	var self = this;

	this.hosts = HostService.query();
	this.monitors = MonitorService.query();
	this.uptime = 0;

	this.agents = {};
	$http.get('/api/agent/').then(function(response) {
		self.agents = response.data;
	});

	this.getHost = function(id) {
		found = {name: '-'};

		self.hosts.forEach(function(host, index) {
			if (host.id == id) {
				found = host;
			}
		});

		return found;
	};

	/**
	 * @expose
	 */
	this.addHost = function() {
		var modalInstance = $uibModal.open({
			animation: true,
			templateUrl: 'newHostModalTemplate',
			controller: 'NewHostController',
			size: 'lg',
			resolve: {
				}
			});
		modalInstance.result.then(function(result) {
			HostService.save(result);
		});
	};

	/**
	 * @expose
	 */
	this.deleteHost = function(id) {
		HostService.delete({id: id});
	};

	/**
	 * @expose
	 */
	this.deleteMonitor = function(id) {
		MonitorService.delete({id: id});
	};

	/**
	 * @expose
	 * @param {string} agentId
	 */
	this.addMonitor = function(agentId) {
		var modalInstance = $uibModal.open({
			animation: true,
			templateUrl: 'newMonitorModalTemplate',
			controller: 'NewMonitorController',
			size: 'lg',
			resolve: {
				agent: function() {
					var agent = self.agents[agentId];
					agent.agentId = agentId;
					return agent;
					},
				hosts: function() {
					return self.hosts;
					}
				}
			});
		modalInstance.result.then(function(result) {
			// Convert to seconds
			result.interval *= 1000000000;
			result.agent.agentId = agentId;
			MonitorService.save(result);
		});
	};

	// Construct websocket url
	var url = '';
	if (window.location.protocol.replace(/:/g, '') === 'https')
		url += 'wss://';
	else
		url += 'ws://';
	url += window.location.host + '/api/ws';

	var socket = new WebSocket(url);

	socket.onmessage = function(msg) {
		var message = JSON.parse(msg.data);
		$scope.$apply(function() {

			switch (message.type) {
				case 'status':
					self.uptime = message.payload.uptime;
					break;
				case 'hostadd':
					self.hosts.push(message.payload);
					break;
				case 'hostchange':
					self.hosts.forEach(function(monitor, index) {
						if (monitor.id == message.payload.id) {
							self.hosts[index] = message.payload;
						}
					});
					break;
				case 'hostdelete':
					self.hosts.forEach(function(monitor, index) {
						if (monitor.id == message.payload) {
							self.hosts.splice(index, 1);
						}
					});
					break;
				case 'monadd':
					self.monitors.push(message.payload);
					break;
				case 'monchange':
					self.monitors.forEach(function(monitor, index) {
						if (monitor.id == message.payload.id) {
							self.monitors[index] = message.payload;
						}
					});
					break;
				case 'mondelete':
					self.monitors.forEach(function(monitor, index) {
						if (monitor.id == message.payload) {
							self.monitors.splice(index, 1);
						}
					});
					break;
				default:
					console.warn("Unsupported message type: " + message.type);
					break;
			}
		});
	};
};

agento.controller('MainController', Agento.Controller.MainController);

/**
 * @ngInject
 * @constructor
 */
Agento.Controller.NewHostController = function($scope, $uibModalInstance) {
	// FIXME: This is currently hardcoded to the ssh transport. This should be fixed.
	$scope.sshPublicKey = window.sshPublicKey;
	$scope.newHost = {};
	$scope.newHost = {
		transportId: 'ssh-command',
		transport: {
			port: 22,
			username: 'root'
		}
	};

	$scope.ok = function() {
		// FIXME: Hardcoded name
		$scope.newHost.name = $scope.newHost.transport.host;

		$uibModalInstance.close($scope.newHost);
	};

	$scope.cancel = function() {
		$uibModalInstance.dismiss('cancel');
	};
};

agento.controller('NewHostController', Agento.Controller.NewHostController);

/**
 * @ngInject
 * @constructor
 */
Agento.Controller.NewMonitorController = function($scope, $uibModalInstance, agent, hosts) {
	$scope.agent = agent;
	$scope.hosts = hosts;
	$scope.newMonitor = {
		agent: {},
		interval: 30,
		hostId: '000000000000000000000000'
	};

	$scope.ok = function() {
		$uibModalInstance.close($scope.newMonitor);
	};

	$scope.cancel = function() {
		$uibModalInstance.dismiss('cancel');
	};
};

agento.controller('NewMonitorController', Agento.Controller.NewMonitorController);

/**
 * @ngInject
 * @constructor
 */
Agento.Filter.GoDuration = function() {
	var Nanosecond = 1;
	var Microsecond = 1000 * Nanosecond;
	var Millisecond = 1000 * Microsecond;
	var Second = 1000 * Millisecond;
	var Minute = 60 * Second;
	var Hour = 60 * Minute;

	/**
	 * @param {string} value
	 * @return {string}
	 */
	var filter = function(value) {
		if (value == undefined)
			return '-';

		var d = parseInt(value);

		if (d < Microsecond) {
			return d + "ns";
		}

		if (d < Millisecond) {
			return d / Microsecond + "µs";
		}

		if (d < Second) {
			return (d / Millisecond).toFixed(2) + "ms";
		}

		if (d < Minute) {
			return (d / Second).toFixed(2) + "s";
		}

		if (d < Hour) {
			return Math.floor(d / Minute) + "m" + (d % Minute / Second).toFixed(2) + "s";
		}

		return Math.floor(d / Hour) + "h" + Math.floor(d % Hour / Minute) + "m" + Math.floor(d % Minute / Second) + "s";
	};

	return filter;
};

agento.filter('goDuration', Agento.Filter.GoDuration);

/**
 * @ngInject
 * @constructor
 */
Agento.Directive.FlashAnim = function() {
	/**
	 * @param {angular.Scope} scope
	 * @param {angular.JQLite} element
	 * @param {angular.Attributes} attrs
	 * @return {void}
	 */
	return function(scope, element, attrs) {
		scope.$watch('monitor', function() {

			element.addClass('flash');
			setTimeout(function() {
				element.removeClass('flash');
			}, 400);
		});
	};
};

agento.directive('flashAnim', Agento.Directive.FlashAnim);
