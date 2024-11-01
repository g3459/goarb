contract Caller {

    uint internal constant STATE_MASK=0x7fffffff00000000000000000000000000000000000000000000000000000000;
    uint internal constant PID_MASK=0xff000000000000000000000000000000000000000000000000000000;
    uint internal constant DIREC_MASK=0x8000000000000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV2_PID=0x01000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV3_PID=0;
    uint internal constant ALGB_PID=0x02000000000000000000000000000000000000000000000000000000;
    uint internal constant TOKEN0_SEL=0x0dfe168100000000000000000000000000000000000000000000000000000000;
    uint internal constant TOKEN1_SEL=0xd21220a700000000000000000000000000000000000000000000000000000000;
    uint internal constant TRANSFER_SEL=0xa9059cbb00000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV2SLOT_SEL=0x0902f1ac00000000000000000000000000000000000000000000000000000000;
    uint internal constant ALGBSLOT_SEL=0xe76c01e400000000000000000000000000000000000000000000000000000000;
    uint internal constant UNIV3SLOT_SEL=0x3850c7bd00000000000000000000000000000000000000000000000000000000;

    address internal constant owner=0xcE158711d84594b298B5EB0eEb16C3018C253371;
    
    constructor() payable{
        unchecked{require(owner==msg.sender);}
    }

    // function setAddress(address a,bool b) external payable check{
    //     unchecked{whitelist[a]=b;}
    // }

    fallback() external payable check{
        unchecked{
            uint fmp=0x80;
            for(uint i;i<msg.data.length;i+=32){
                uint poolCall;
                assembly{
                    poolCall:=calldataload(i)
                }
                uint pid = poolCall&PID_MASK;
                uint outsize;
                if(pid==UNIV2_PID){
                    outsize=0x40;
                    assembly{mstore(fmp,UNIV2SLOT_SEL)}
                }else if(pid==ALGB_PID){
                    assembly{mstore(fmp,ALGBSLOT_SEL)}
                }else{
                    assembly{mstore(fmp,UNIV3SLOT_SEL)}
                }
                assembly{
                    pop(call(gas(), poolCall, 0, fmp, 0x04, fmp, outsize))
                    if xor(and(keccak256(0x80,0x20),STATE_MASK),and(poolCall,STATE_MASK)){
                        revert(0,0)
                    }
                }
                fmp+=outsize;
            }
            uint fmp2=0x80;
            for(uint i;i<msg.data.length;i+=32){
                uint poolCall;
                assembly{
                    poolCall:=calldataload(i)
                }
                uint amIn=uint(uint48(poolCall>>168))<<uint8(poolCall>>160);
                address pool=address(uint160(poolCall));
                bool direc;
                assembly{direc:=and(poolCall,DIREC_MASK)}
                if(poolCall&PID_MASK==UNIV2_PID){
                    uint rIn;uint rOut;
                    assembly{
                        rIn:=mload(fmp2)
                        rOut:=mload(add(fmp2,0x20))
                    }
                    fmp2+=0x40;
                    if(direc){
                        assembly{mstore(fmp,TOKEN0_SEL)}
                    }else{
                        (rIn,rOut)=(rOut,rIn);
                        assembly{mstore(fmp,TOKEN1_SEL)}
                    }
                    assembly{
                        pop(call(gas(), poolCall, 0, fmp, 0x04, fmp, 0x20))
                        let token:=mload(fmp)
                        mstore(fmp,TRANSFER_SEL)
                        mstore(add(fmp,0x04),pool)
                        mstore(add(fmp,0x24),amIn)
                        pop(call(gas(), token, 0, fmp, 0x44, 0, 0))
                    }
                    uint amOut=(amIn-1)*997;
                    amOut = (amOut * rOut) / (rIn * 1000 + amOut) - 1;
                    IUniV2Pool(pool).swap(direc?0:amOut, direc?amOut:0, address(this), "");
                }else{
                    IUniV3Pool(pool).swap(address(this), direc, int(amIn) , direc ? 4295128740 : 1461446703485210103287273052203988822378723970341, "");
                }
            }
        }
    }
    // receive() external payable{}

    // function recover() external payable check{
    //     unchecked{payable(msg.sender).transfer(address(this).balance);}
    // }

    function execute(address target, bytes calldata call) external payable check returns (bool s){
        unchecked{(s,)=target.call(call);}
    }

    modifier check{
        _;
        unchecked{require(owner==tx.origin || owner==msg.sender);}
    }

    modifier swapCallback(int am0,int am1){
        _;
        unchecked{
            uint amIn;
            if(am0>am1){
                amIn=uint(am0);
                assembly{mstore(0x80,TOKEN0_SEL)}
            }else{
                amIn=uint(am1);
                assembly{mstore(0x80,TOKEN1_SEL)}
            }
            assembly{
                pop(call(gas(), caller(), 0, 0x80, 0x04, 0x80, 0x20))
                let token:=mload(0x80)
                mstore(0x80,TRANSFER_SEL)
                mstore(0x84,caller())
                mstore(0xa4,amIn)
                pop(call(gas(), token, 0, 0x80, 0x44, 0, 0))
            }    
        }
    }

    function uniswapV3SwapCallback(int am0 , int am1, bytes calldata) external payable swapCallback(am0,am1) check{}

    function pancakeV3SwapCallback(int am0 , int am1, bytes calldata) external payable swapCallback(am0,am1) check{}

    function algebraSwapCallback(int am0, int am1, bytes calldata) external payable swapCallback(am0,am1) check{}

}

interface IERC20{
    function balanceOf(address ) external view returns ( uint256 );
    function transfer(address, uint256 ) external returns ( uint256 );
}

interface IUniV3Pool{
    function token0() external view returns ( address );
    function token1() external view returns ( address );
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function slot0() external view returns(uint sqrtPX96, int t, uint observationIndex, uint observationCardinality, uint observationCardinalityNext, uint feeProtocol, bool unlocked);
}

interface IUniV2Pool{
    function swap(uint amount0Out, uint amount1Out, address to, bytes calldata data) external;
    function getReserves()external view returns(uint reserve0, uint reserve1, uint blockTimestampLast);
    function token0() external view returns ( address );
    function token1() external view returns ( address );
}

interface IAlgebraV3Pool{
    function swap(address recipient, bool zeroForOne, int amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data) external returns(int amount0, int amount1);
    function globalState() external view returns(uint sqrtPX96, int t, uint fee, uint timePointIndex, uint comunityFeet0, uint comunityFeeT1, bool unlocked);
    function tickSpacing() external view returns(int s);
    function liquidity() external view returns(uint liquidity);
}