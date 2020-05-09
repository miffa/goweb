# Generated by the protocol buffer compiler.  DO NOT EDIT!
# Source: grpclb_balancer_v1/balancer.proto for package 'grpclb.balancer.v1'

require 'grpc'
require 'grpclb_balancer_v1/balancer_pb'

module Grpclb
  module Balancer
    module V1
      module LoadBalancer
        class Service

          include GRPC::GenericService

          self.marshal_class_method = :encode
          self.unmarshal_class_method = :decode
          self.service_name = 'grpclb.balancer.v1.LoadBalancer'

          rpc :Servers, ServersRequest, ServersResponse
        end

        Stub = Service.rpc_stub_class
      end
    end
  end
end
