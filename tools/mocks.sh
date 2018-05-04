#!/bin/bash

mockery -case=underscore -name Application -output=mocks -outpkg=communication_mocks
