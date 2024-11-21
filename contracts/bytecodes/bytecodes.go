package bytecodes

import (
	"encoding/hex"
)

var RouterBytecode, _ = hex.DecodeString("6080806040523460195761101b908161001e823930815050f35b5f80fdfe60806040526004361015610011575f80fd5b5f3560e01c633818b99d14610024575f80fd5b60807ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffc3601126101ee576100566101f7565b61005e610207565b906064359067ffffffffffffffff82116101ee57366023830112156101ee5781600401359261008c84610285565b9261009a6040519485610244565b8484526024602085019560051b820101903682116101ee5760248101955b8287106100e1576100cd866044358688610428565b906100dd6040519283928361029d565b0390f35b863567ffffffffffffffff81116101ee578201366043820112156101ee57602481013561010d81610285565b9161011b6040519384610244565b818352602060248185019360051b83010101903682116101ee5760448101925b8284106101555750505090825250602096870196016100b8565b833567ffffffffffffffff81116101ee5760249083010136603f820112156101ee5760208101359167ffffffffffffffff83116101f2576040516101c160207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f8701160182610244565b83815236604084860101116101ee575f60208581966040839701838601378301015281520193019261013b565b5f80fd5b610217565b6004359060ff821682036101ee57565b6024359060ff821682036101ee57565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b90601f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0910116810190811067ffffffffffffffff8211176101f257604052565b67ffffffffffffffff81116101f25760051b60200190565b604081016040825282518091526020606083019301905f5b818110610370575050506020818303910152815180825260208201916020808360051b8301019401925f915b8383106102f057505050505090565b909192939460208080837fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe086600196030187527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f838c518051918291828752018686015e5f85828601015201160101970193019301919392906102e1565b82518552602094850194909201916001016102b5565b9061039082610285565b61039d6040519182610244565b8281527fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe06103cb8294610285565b0190602036910137565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b80511561040f5760200190565b6103d5565b805182101561040f5760209160051b010190565b9290939161044360ff61043b8451610386565b961686610414565b528051907fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe061048a61047484610285565b936104826040519586610244565b808552610285565b015f5b8181106108625750509060ff819460051b16906104aa8351610386565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8451610100031c805b6104df575050505050565b909192945f97949697505f915b8751831015610850576001831b908181161561084357189061050e8387610414565b5115801561082f575b610824575f915b885183101561081257828414610809575f610539858b610414565b51511515806107f1575b156107a25750600161055f84610559878d610414565b51610414565b515b8461078c5785896105af8a5f945b86866105a8877fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6105a0828a610414565b510195610414565b51936108e0565b936105ba8984610414565b5182111561077d578a6105cd8582610414565b516105f57bff0000000000000000000000000000000000000000000000000000008816610b5b565b01913a8302903a6106068d85610414565b5102908061075f575b5061061a8c87610414565b5103908403131561074f576106a261069c6106d49660019a99968d6106967fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6107059c997f7fffffffff00000000000000ffffffffffffffffffffffffffffffffffffffff9961068d856106a89b610414565b52019183610414565b52610414565b51610b8b565b60a01b90565b91161790610728575b6107006106be898c610414565b5191604051938491602083019190602083019252565b037fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08101845283610244565b6108bc565b61070f8589610414565b5261071a8488610414565b5081841b17925b019161051e565b7f8000000000000000000000000000000000000000000000000000000000000000176106b1565b5050505050505091600190610721565b61076f8161077793948802610873565b928602610873565b5f61060f565b50505050505091600190610721565b85896105af8a61079b83610402565b519461056f565b6107ac848b610414565b51511515806107d9575b156107cf576107c985610559868d610414565b51610561565b5091600190610721565b506107e885610559868d610414565b515115156107b6565b5061080084610559878d610414565b51511515610543565b91600190610721565b926001919992505b01919790976104ec565b97909160019061081a565b508661083b8487610414565b515114610517565b926001915098919861081a565b949796909590949093929150806104d4565b80606060208093870101520161048d565b811561087d570490565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b805191908290602001825e015f815290565b6108de906106d46108d8949360405195869360208501906108aa565b906108aa565b565b92949392915f908180805b8951821015610b4f57818a019460408601519561091e73ffffffffffffffffffffffffffffffffffffffff88168b610bb1565b610b435760200151936fffffffffffffffffffffffffffffffff6109428660801c90565b951692838715610b3a575b5061096561095e61095e8a60a01c90565b61ffff1690565b80620f42400390818c0290610985620f42408a0292888185019102610873565b988d878b1115610a4d57908b918b8e7bff0000000000000000000000000000000000000000000000000000008116957b010000000000000000000000000000000000000000000000000000008703610a74575b5050505050506109e83a91610b5b565b02958b80610a61575b50868903928d868803851315610a4d57928892610a2092610a15610a289660011b90565b028092019102610873565b039160011b90565b12610a3e575050506040909294915b01906108eb565b93919650935060409150610a37565b505050505093919650935060409150610a37565b610a6d91978a02610873565b958b6109f1565b610a8a610a84610abd9360b01c90565b60020b90565b907b020000000000000000000000000000000000000000000000000000008814610b3157610ab790610c0a565b90610c5c565b909415610b0d575091610ae191610ad9610aea948d0360801b90565b910190610873565b9160020b610cdf565b115b610afb57895f8e8b8e836109d8565b50505093919650935060409150610a37565b935050610b22610ae191610b2b930160801b90565b8c8b0390610873565b10610aec565b50603c90610c5c565b9593505f61094d565b50945090604090610a37565b97965050505094505050565b7b0200000000000000000000000000000000000000000000000000000003610b8457620493e090565b620186a090565b5f905b8065ffffffffffff811603610ba45760081b1790565b906008019060081c610b8e565b9060205b82518111610c035773ffffffffffffffffffffffffffffffffffffffff818401511673ffffffffffffffffffffffffffffffffffffffff831614610bfb57602001610bb5565b505050600190565b5050505f90565b905f9160648114610c55576101f48114610c4e576109c48114610c4757610bb88114610c405761271014610c3a57565b60c89150565b50603c9150565b5060329150565b50600a9150565b5060019150565b8162ffffff91818082075f83121691050302911660020b810160020b907ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff276188160020b125f14610cca57507ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2761891565b91620d89e88213610cd757565b620d89e89150565b610de2908060ff1d8181011890600182167001fffcb933bd6fad37aa2d162d1a59400102700100000000000000000000000000000000189160028116610fff575b60048116610fe3575b60088116610fc7575b60108116610fab575b60208116610f8f575b60408116610f73575b60808116610f57575b6101008116610f3b575b6102008116610f1f575b6104008116610f03575b6108008116610ee7575b6110008116610ecb575b6120008116610eaf575b6140008116610e93575b6180008116610e77575b620100008116610e5b575b620200008116610e40575b620400008116610e25575b6208000016610e0c575b5f12610de5575b60401c6002900a90565b90565b7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04610dd8565b6b048a170391f7dc42444e8fa290910260801c90610dd1565b6d2216e584f5fa1ea926041bedfe9890920260801c91610dc7565b916e5d6af8dedb81196699c329225ee6040260801c91610dbc565b916f09aa508b5b7a84e1c677de54f3e99bc90260801c91610db1565b916f31be135f97d08fd981231505542fcfa60260801c91610da6565b916f70d869a156d2a1b890bb3df62baf32f70260801c91610d9c565b916fa9f746462d870fdf8a65dc1f90e061e50260801c91610d92565b916fd097f3bdfd2022b8845ad8f792aa58250260801c91610d88565b916fe7159475a2c29b7443b29c7fa6e889d90260801c91610d7e565b916ff3392b0822b70005940c7a398e4b70f30260801c91610d74565b916ff987a7253ac413176f2b074cf7815e540260801c91610d6a565b916ffcbe86c7900a88aedcffc83b479aa3a40260801c91610d60565b916ffe5dee046a99a2a811c461f1969c30530260801c91610d56565b916fff2ea16466c96a3843ec78b326b528610260801c91610d4d565b916fff973b41fa98c081472e6896dfb254c00260801c91610d44565b916fffcb9843d60f6159c9db58835c9266440260801c91610d3b565b916fffe5caca7e10e4e61c3624eaa0941cd00260801c91610d32565b916ffff2e50f5f656932ef12357cf3c7fdcc0260801c91610d29565b916ffff97272373d413259a46990580e213a0260801c91610d2056")
