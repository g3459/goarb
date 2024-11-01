package bytecodes

import (
	"encoding/hex"
)

var RouterBytecode, _ = hex.DecodeString("60808060405234601957611066908161001e823930815050f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c633818b99d14610024575f80fd5b60807ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3601126101ee576100566101f7565b61005e610207565b906064359067ffffffffffffffff82116101ee57366023830112156101ee5781600401359261008c84610285565b9261009a6040519485610244565b8484526024602085019560051b820101903682116101ee5760248101955b8287106100e1576100cd866044358688610428565b906100dd6040519283928361029d565b0390f35b863567ffffffffffffffff81116101ee578201366043820112156101ee57602481013561010d81610285565b9161011b6040519384610244565b818352602060248185019360051b83010101903682116101ee5760448101925b8284106101555750505090825250602096870196016100b8565b833567ffffffffffffffff81116101ee5760249083010136603f820112156101ee5760208101359167ffffffffffffffff83116101f2576040516101c160207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f8701160182610244565b83815236604084860101116101ee575f60208581966040839701838601378301015281520193019261013b565b5f80fd5b610217565b6004359060ff821682036101ee57565b6024359060ff821682036101ee57565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b90601f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0910116810190811067ffffffffffffffff8211176101f257604052565b67ffffffffffffffff81116101f25760051b60200190565b604081016040825282518091526020606083019301905f5b818110610370575050506020818303910152815180825260208201916020808360051b8301019401925f915b8383106102f057505050505090565b909192939460208080837fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe086600196030187527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f838c518051918291828752018686015e5f85828601015201160101970193019301919392906102e1565b82518552602094850194909201916001016102b5565b9061039082610285565b61039d6040519182610244565b8281527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe06103cb8294610285565b0190602036910137565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b80511561040f5760200190565b6103d5565b805182101561040f5760209160051b010190565b9290939161044360ff61043b8451610386565b961686610414565b528051907fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe061048a61047484610285565b936104826040519586610244565b808552610285565b015f5b81811061084357505060ff829460051b166104a88251610386565b927fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8351610100031c805b6104de575050505050565b91945f979491939697505f925b8751841015610832576001841b908181161561082557189061050d8487610414565b51158015610811575b610804575f915b88518310156107f1578285146107e8575f610538868b610414565b51511515806107d0575b156107795750888486896105628761055c84600197610414565b51610414565b515b87610765576105b08b5f925b87846105a9887fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6105a1828b610414565b510195610414565b51936108c1565b94906105bc8583610414565b516105e47bff0000000000000000000000000000000000000000000000000000008816610b3c565b01923a8402903a6105f58d86610414565b51029080610747575b506106098c87610414565b510390820313156107375761068a6106846106bc9660019a99966106ed999661067e8f7f7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff987fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff610690990161067e8387610414565b52610414565b51610b6c565b60a01b90565b91161790610710575b6106e86106a68a8c610414565b5191604051938491602083019190602083019252565b037fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08101845283610244565b61089d565b6106f78589610414565b526107028488610414565b5081841b17925b019161051d565b7f800000000000000000000000000000000000000000000000000000000000000017610699565b5050505050505091600190610709565b6107578161075f93948602610854565b928402610854565b5f6105fe565b6105b08b61077284610402565b5192610570565b8486898c6107878882610414565b51511515806107b8575b156107aa578261055c896107a493610414565b51610564565b505050505091600190610709565b506107c78361055c8a84610414565b51511515610791565b506107df8461055c888d610414565b51511515610542565b91600190610709565b92989150926001905b01929097916104eb565b92600190989192986107fa565b508661081d8587610414565b515114610516565b92989193600191506107fa565b8095989794929196935090946104d3565b80606060208093870101520161048d565b811561085e570490565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b805191908290602001825e015f815290565b6108bf906106bc6108b99493604051958693602085019061088b565b9061088b565b565b92949392915f908180805b8951821015610b3057818a01946040860151956108ff73ffffffffffffffffffffffffffffffffffffffff88168b610bbf565b610b245760200151936fffffffffffffffffffffffffffffffff6109238660801c90565b951692838715610b1b575b5061094661093f61093f8a60a01c90565b61ffff1690565b80620f42400390818c0290610966620f42408a0292888185019102610854565b988d878b1115610a2e57908b918b8e7bff0000000000000000000000000000000000000000000000000000008116957b010000000000000000000000000000000000000000000000000000008703610a55575b5050505050506109c93a91610b3c565b02958b80610a42575b50868903928d868803851315610a2e57928892610a01926109f6610a099660011b90565b028092019102610854565b039160011b90565b12610a1f575050506040909294915b01906108cc565b93919650935060409150610a18565b505050505093919650935060409150610a18565b610a4e91978a02610854565b958b6109d2565b610a6b610a65610a9e9360b01c90565b60020b90565b907b020000000000000000000000000000000000000000000000000000008814610b1257610a9890610c23565b90610c75565b909415610aee575091610ac291610aba610acb948d0360801b90565b910190610854565b9160020b610d2a565b115b610adc57895f8e8b8e836109b9565b50505093919650935060409150610a18565b935050610b03610ac291610b0c930160801b90565b8c8b0390610854565b10610acd565b50603c90610c75565b9593505f61092e565b50945090604090610a18565b97965050505094505050565b7b0200000000000000000000000000000000000000000000000000000003610b6557620493e090565b620186a090565b5f905b8065ffffffffffff811603610b855760081b1790565b906008019060081c610b6f565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b919060205b83518111610c1c5773ffffffffffffffffffffffffffffffffffffffff818501511673ffffffffffffffffffffffffffffffffffffffff831614610c14576020810180911115610bc4575b610b92565b506001925050565b505f925050565b905f9160648114610c6e576101f48114610c67576109c48114610c6057610bb88114610c595761271014610c5357565b60c89150565b50603c9150565b5060329150565b50600a9150565b5060019150565b8190818082075f8312169105030262ffffff8160020b921660020b82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8000008112627fffff821317610c0f577ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff276188193125f14610d105750507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2761891565b620d89e89060029492940b13610d2257565b620d89e89150565b610e2d908060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a5940010270010000000000000000000000000000000018916002811661104a575b6004811661102e575b60088116611012575b60108116610ff6575b60208116610fda575b60408116610fbe575b60808116610fa2575b6101008116610f86575b6102008116610f6a575b6104008116610f4e575b6108008116610f32575b6110008116610f16575b6120008116610efa575b6140008116610ede575b6180008116610ec2575b620100008116610ea6575b620200008116610e8b575b620400008116610e70575b6208000016610e57575b5f12610e30575b60401c6002900a90565b90565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04610e23565b6b048a170391f7dc42444e8fa290910260801c90610e1c565b6d2216e584f5fa1ea926041bedfe9890920260801c91610e12565b916e5d6af8dedb81196699c329225ee6040260801c91610e07565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610dfc565b916f31be135f97d08fd981231505542fcfa60260801c91610df1565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610de7565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610ddd565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610dd3565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610dc9565b916ff3392b0822b70005940c7a398e4b70f30260801c91610dbf565b916ff987a7253ac413176f2b074cf7815e540260801c91610db5565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610dab565b916ffe5dee046a99a2a811c461f1969c30530260801c91610da1565b916fff2ea16466c96a3843ec78b326b528610260801c91610d98565b916fff973b41fa98c081472e6896dfb254c00260801c91610d8f565b916fffcb9843d60f6159c9db58835c9266440260801c91610d86565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610d7d565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610d74565b916ffff97272373d413259a46990580e213a0260801c91610d6b56")
