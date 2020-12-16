/**
 * @fileoverview gRPC-Web generated client stub for pb
 * @enhanceable
 * @public
 */

// GENERATED CODE -- DO NOT EDIT!


/* eslint-disable */
// @ts-nocheck



const grpc = {};
grpc.web = require('grpc-web');


var models_pb = require('./models_pb.js')
const proto = {};
proto.pb = require('./api_pb.js');

/**
 * @param {string} hostname
 * @param {?Object} credentials
 * @param {?Object} options
 * @constructor
 * @struct
 * @final
 */
proto.pb.TRISADirectoryClient =
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
proto.pb.TRISADirectoryPromiseClient =
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
 *   !proto.pb.RegisterRequest,
 *   !proto.pb.RegisterReply>}
 */
const methodDescriptor_TRISADirectory_Register = new grpc.web.MethodDescriptor(
  '/pb.TRISADirectory/Register',
  grpc.web.MethodType.UNARY,
  proto.pb.RegisterRequest,
  proto.pb.RegisterReply,
  /**
   * @param {!proto.pb.RegisterRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.pb.RegisterReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.pb.RegisterRequest,
 *   !proto.pb.RegisterReply>}
 */
const methodInfo_TRISADirectory_Register = new grpc.web.AbstractClientBase.MethodInfo(
  proto.pb.RegisterReply,
  /**
   * @param {!proto.pb.RegisterRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.pb.RegisterReply.deserializeBinary
);


/**
 * @param {!proto.pb.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.pb.RegisterReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.pb.RegisterReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.pb.TRISADirectoryClient.prototype.register =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/pb.TRISADirectory/Register',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Register,
      callback);
};


/**
 * @param {!proto.pb.RegisterRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.pb.RegisterReply>}
 *     A native promise that resolves to the response
 */
proto.pb.TRISADirectoryPromiseClient.prototype.register =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/pb.TRISADirectory/Register',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Register);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.pb.LookupRequest,
 *   !proto.pb.LookupReply>}
 */
const methodDescriptor_TRISADirectory_Lookup = new grpc.web.MethodDescriptor(
  '/pb.TRISADirectory/Lookup',
  grpc.web.MethodType.UNARY,
  proto.pb.LookupRequest,
  proto.pb.LookupReply,
  /**
   * @param {!proto.pb.LookupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.pb.LookupReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.pb.LookupRequest,
 *   !proto.pb.LookupReply>}
 */
const methodInfo_TRISADirectory_Lookup = new grpc.web.AbstractClientBase.MethodInfo(
  proto.pb.LookupReply,
  /**
   * @param {!proto.pb.LookupRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.pb.LookupReply.deserializeBinary
);


/**
 * @param {!proto.pb.LookupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.pb.LookupReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.pb.LookupReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.pb.TRISADirectoryClient.prototype.lookup =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/pb.TRISADirectory/Lookup',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Lookup,
      callback);
};


/**
 * @param {!proto.pb.LookupRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.pb.LookupReply>}
 *     A native promise that resolves to the response
 */
proto.pb.TRISADirectoryPromiseClient.prototype.lookup =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/pb.TRISADirectory/Lookup',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Lookup);
};


/**
 * @const
 * @type {!grpc.web.MethodDescriptor<
 *   !proto.pb.SearchRequest,
 *   !proto.pb.SearchReply>}
 */
const methodDescriptor_TRISADirectory_Search = new grpc.web.MethodDescriptor(
  '/pb.TRISADirectory/Search',
  grpc.web.MethodType.UNARY,
  proto.pb.SearchRequest,
  proto.pb.SearchReply,
  /**
   * @param {!proto.pb.SearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.pb.SearchReply.deserializeBinary
);


/**
 * @const
 * @type {!grpc.web.AbstractClientBase.MethodInfo<
 *   !proto.pb.SearchRequest,
 *   !proto.pb.SearchReply>}
 */
const methodInfo_TRISADirectory_Search = new grpc.web.AbstractClientBase.MethodInfo(
  proto.pb.SearchReply,
  /**
   * @param {!proto.pb.SearchRequest} request
   * @return {!Uint8Array}
   */
  function(request) {
    return request.serializeBinary();
  },
  proto.pb.SearchReply.deserializeBinary
);


/**
 * @param {!proto.pb.SearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @param {function(?grpc.web.Error, ?proto.pb.SearchReply)}
 *     callback The callback function(error, response)
 * @return {!grpc.web.ClientReadableStream<!proto.pb.SearchReply>|undefined}
 *     The XHR Node Readable Stream
 */
proto.pb.TRISADirectoryClient.prototype.search =
    function(request, metadata, callback) {
  return this.client_.rpcCall(this.hostname_ +
      '/pb.TRISADirectory/Search',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Search,
      callback);
};


/**
 * @param {!proto.pb.SearchRequest} request The
 *     request proto
 * @param {?Object<string, string>} metadata User defined
 *     call metadata
 * @return {!Promise<!proto.pb.SearchReply>}
 *     A native promise that resolves to the response
 */
proto.pb.TRISADirectoryPromiseClient.prototype.search =
    function(request, metadata) {
  return this.client_.unaryCall(this.hostname_ +
      '/pb.TRISADirectory/Search',
      request,
      metadata || {},
      methodDescriptor_TRISADirectory_Search);
};


module.exports = proto.pb;

