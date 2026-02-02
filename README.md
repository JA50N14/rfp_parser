1.Point the const variable "RfpPackageRootDir" in main.go to a directory containing RFP Packages for a specific Business Unit.
2.Each sub-folder in the "RfpPackageRootDir" path must represent a single RFP Package that was issued by the potential client in the same year.
3.Run the program with: go run ./ Year BusinessUnit
    -Example: go run ./ 2025 ABS
