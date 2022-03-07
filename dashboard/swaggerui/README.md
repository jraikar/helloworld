# Using Aerostation Swagger


* Quick testing of swagger specs (follow steps [1-2])
* For local testing and updating swagger spec (follow steps [1-4]) 
* Commands to generate spec and serve are also available in Makefile


1. Install swagger cli - https://goswagger.io/install.html
2. Server specification UI
    - `swagger serve https://<github.static.url.to.swagger>.json`
    - or
    - From source file - `swagger serve dashboard/swaggerui/swagger.json`

3. [Validate](https://goswagger.io/usage/validate.html) specification
    - `swagger validate <path_to_swagger_spec.json>`
    
4. [Generate](https://goswagger.io/generate/server.html) a spec from source
    - `swagger generate spec -o dashboard/swaggerui/swagger.json -m`
   