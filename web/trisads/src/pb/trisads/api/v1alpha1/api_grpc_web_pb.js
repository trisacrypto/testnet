/**
 * @fileoverview gRPC-Web generated client stub for trisads.api.v1alpha1
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var ivms101_ivms101_pb = require('../../../ivms101/ivms101_pb.js')

var trisads_models_v1alpha1_models_pb = require('../../../trisads/models/v1alpha1/models_pb.js')

var trisads_models_v1alpha1_ca_pb = require('../../../trisads/models/v1alpha1/ca_pb.js')
const proto = {};
proto.trisads = {};
proto.trisads.api = {};
proto.trisads.api.v1alpha1 = require('./api_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.trisads.api.v1alpha1.TRISADirectoryClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.trisads.api.v1alpha1.TRISADirectoryPromiseClient =
    function(hostname, credentials, options) {
  if (!options) options = {};
  options['format'] = 'text';

  /**
   * @private @const {!grpc.web.GrpcWebClientBase} The client
   */
  this.client_ = new grpc.web.GrpcWebClientBase(options);

  /**
   * @private @const {string} The hostname
   */
  this.hostname_ = hostname;

};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.trisads.api.v1alpha1.RegisterRequest,
 *   !proto.trisads.api.v1alpha1.RegisterReply>}
 */
const methodDescriptor_TRISADirectory_Register = new grpc.web.MethodDescriptor(
  '/trisads.api.v1alpha1.TRISADirectory/Register',
  grpc.web.MethodType.UNARY,
  proto.trisads.api.v1alpha1.RegisterRequest,
  proto.trisads.api.v1alpha1.RegisterReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.RegisterRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.RegisterReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.trisads.api.v1alpha1.RegisterRequest,
 *   !proto.trisads.api.v1alpha1.RegisterReply>}
 */
const methodInfo_TRISADirectory_Register = new grpc.web.AbstractClientBase.MethodInfo(
  proto.trisads.api.v1alpha1.RegisterReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.RegisterRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.RegisterReply.deserializeBinary
);


/**
 * @param {!proto.trisads.api.v1alpha1.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.trisads.api.v1alpha1.RegisterReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.trisads.api.v1alpha1.RegisterReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.trisads.api.v1alpha1.TRISADirectoryClient.prototype.register =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Register',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Register,
      callback);
};


/**
 * @param {!proto.trisads.api.v1alpha1.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.trisads.api.v1alpha1.RegisterReply>}
 *     Promise that resolves to the response
 */
proto.trisads.api.v1alpha1.TRISADirectoryPromiseClient.prototype.register =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Register',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Register);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.trisads.api.v1alpha1.LookupRequest,
 *   !proto.trisads.api.v1alpha1.LookupReply>}
 */
const methodDescriptor_TRISADirectory_Lookup = new grpc.web.MethodDescriptor(
  '/trisads.api.v1alpha1.TRISADirectory/Lookup',
  grpc.web.MethodType.UNARY,
  proto.trisads.api.v1alpha1.LookupRequest,
  proto.trisads.api.v1alpha1.LookupReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.LookupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.LookupReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.trisads.api.v1alpha1.LookupRequest,
 *   !proto.trisads.api.v1alpha1.LookupReply>}
 */
const methodInfo_TRISADirectory_Lookup = new grpc.web.AbstractClientBase.MethodInfo(
  proto.trisads.api.v1alpha1.LookupReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.LookupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.LookupReply.deserializeBinary
);


/**
 * @param {!proto.trisads.api.v1alpha1.LookupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.trisads.api.v1alpha1.LookupReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.trisads.api.v1alpha1.LookupReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.trisads.api.v1alpha1.TRISADirectoryClient.prototype.lookup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Lookup',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Lookup,
      callback);
};


/**
 * @param {!proto.trisads.api.v1alpha1.LookupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.trisads.api.v1alpha1.LookupReply>}
 *     Promise that resolves to the response
 */
proto.trisads.api.v1alpha1.TRISADirectoryPromiseClient.prototype.lookup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Lookup',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Lookup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.trisads.api.v1alpha1.SearchRequest,
 *   !proto.trisads.api.v1alpha1.SearchReply>}
 */
const methodDescriptor_TRISADirectory_Search = new grpc.web.MethodDescriptor(
  '/trisads.api.v1alpha1.TRISADirectory/Search',
  grpc.web.MethodType.UNARY,
  proto.trisads.api.v1alpha1.SearchRequest,
  proto.trisads.api.v1alpha1.SearchReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.SearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.SearchReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.trisads.api.v1alpha1.SearchRequest,
 *   !proto.trisads.api.v1alpha1.SearchReply>}
 */
const methodInfo_TRISADirectory_Search = new grpc.web.AbstractClientBase.MethodInfo(
  proto.trisads.api.v1alpha1.SearchReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.SearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.SearchReply.deserializeBinary
);


/**
 * @param {!proto.trisads.api.v1alpha1.SearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.trisads.api.v1alpha1.SearchReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.trisads.api.v1alpha1.SearchReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.trisads.api.v1alpha1.TRISADirectoryClient.prototype.search =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Search',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Search,
      callback);
};


/**
 * @param {!proto.trisads.api.v1alpha1.SearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.trisads.api.v1alpha1.SearchReply>}
 *     Promise that resolves to the response
 */
proto.trisads.api.v1alpha1.TRISADirectoryPromiseClient.prototype.search =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Search',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Search);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.trisads.api.v1alpha1.StatusRequest,
 *   !proto.trisads.api.v1alpha1.StatusReply>}
 */
const methodDescriptor_TRISADirectory_Status = new grpc.web.MethodDescriptor(
  '/trisads.api.v1alpha1.TRISADirectory/Status',
  grpc.web.MethodType.UNARY,
  proto.trisads.api.v1alpha1.StatusRequest,
  proto.trisads.api.v1alpha1.StatusReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.StatusRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.StatusReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.trisads.api.v1alpha1.StatusRequest,
 *   !proto.trisads.api.v1alpha1.StatusReply>}
 */
const methodInfo_TRISADirectory_Status = new grpc.web.AbstractClientBase.MethodInfo(
  proto.trisads.api.v1alpha1.StatusReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.StatusRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.StatusReply.deserializeBinary
);


/**
 * @param {!proto.trisads.api.v1alpha1.StatusRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.trisads.api.v1alpha1.StatusReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.trisads.api.v1alpha1.StatusReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.trisads.api.v1alpha1.TRISADirectoryClient.prototype.status =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Status',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Status,
      callback);
};


/**
 * @param {!proto.trisads.api.v1alpha1.StatusRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.trisads.api.v1alpha1.StatusReply>}
 *     Promise that resolves to the response
 */
proto.trisads.api.v1alpha1.TRISADirectoryPromiseClient.prototype.status =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/Status',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Status);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.trisads.api.v1alpha1.VerifyEmailRequest,
 *   !proto.trisads.api.v1alpha1.VerifyEmailReply>}
 */
const methodDescriptor_TRISADirectory_VerifyEmail = new grpc.web.MethodDescriptor(
  '/trisads.api.v1alpha1.TRISADirectory/VerifyEmail',
  grpc.web.MethodType.UNARY,
  proto.trisads.api.v1alpha1.VerifyEmailRequest,
  proto.trisads.api.v1alpha1.VerifyEmailReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.VerifyEmailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.VerifyEmailReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.trisads.api.v1alpha1.VerifyEmailRequest,
 *   !proto.trisads.api.v1alpha1.VerifyEmailReply>}
 */
const methodInfo_TRISADirectory_VerifyEmail = new grpc.web.AbstractClientBase.MethodInfo(
  proto.trisads.api.v1alpha1.VerifyEmailReply,
  /**
   * @param {!proto.trisads.api.v1alpha1.VerifyEmailRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.trisads.api.v1alpha1.VerifyEmailReply.deserializeBinary
);


/**
 * @param {!proto.trisads.api.v1alpha1.VerifyEmailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.trisads.api.v1alpha1.VerifyEmailReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.trisads.api.v1alpha1.VerifyEmailReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.trisads.api.v1alpha1.TRISADirectoryClient.prototype.verifyEmail =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/VerifyEmail',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_VerifyEmail,
      callback);
};


/**
 * @param {!proto.trisads.api.v1alpha1.VerifyEmailRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.trisads.api.v1alpha1.VerifyEmailReply>}
 *     Promise that resolves to the response
 */
proto.trisads.api.v1alpha1.TRISADirectoryPromiseClient.prototype.verifyEmail =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/trisads.api.v1alpha1.TRISADirectory/VerifyEmail',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_VerifyEmail);
};


module.exports = proto.trisads.api.v1alpha1;

