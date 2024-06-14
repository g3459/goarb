
contract Caller {

    bool lock;
    mapping(address=>bool) public whitelist;

    constructor() payable{
        unchecked{
            whitelist[msg.sender]=true;
        }
    }

    function setAddress(address a,bool b) external payable locked{
        unchecked{
            whitelist[a]=b;
        }
    }

    fallback() external payable locked{
        executeRoute(msg.data);
    }

    receive() external payable{}

    function recover() external payable locked{
        payable(msg.sender).transfer(address(this).balance);
    }

    function executeRoute(bytes calldata calls) internal{
        unchecked{
            bytes32 poolCall=bytes32(calls[calls.length-32:]);
            address pool;
            assembly{pool:=poolCall}
            uint t=uint8(uint(poolCall)>>216);
            (, bytes memory state)=pool.call(abi.encodeWithSelector(t==0?IUniV3Pool.slot0.selector:(t==1?IUniV2Pool.getReserves.selector:IAlgebraV3Pool.globalState.selector)));
            require(bytes4(keccak256(state))<<1==bytes4(poolCall)<<1,"1");
            if(calls.length>32)
                executeRoute(calls[:calls.length-32]);
            bool direc=bytes1(poolCall)&bytes1(0x80)==bytes1(0x80);
            uint amIn=uint(uint48(uint(poolCall)>>168))<<uint8(uint(poolCall)>>160);
            if(t==0 || t==2){
                IUniV3Pool(pool).swap(address(this), direc, int(amIn) , direc ? 4295128740 : 1461446703485210103287273052203988822378723970341, "");
            }else{
                (uint reserve0, uint reserve1)=abi.decode(state,(uint,uint));
                uint amOut=(amIn-1)*997000;
                amOut = (direc
                    ? (amOut * reserve1) / (reserve0 * 1e6 + amOut)
                    : (amOut * reserve0) / (reserve1 * 1e6 + amOut))-1;
                IERC20(direc?IUniV2Pool(pool).token0():IUniV2Pool(pool).token1()).transfer(pool,amIn);
                IUniV2Pool(pool).swap(direc?0:amOut, direc?amOut:0, address(this), "");
            }
        }
    }

    function execute(address target, bytes calldata call) public payable locked returns (bool s){
        unchecked{
            (s,)=target.call(call);
        }
    }

    modifier locked{
        if(lock){
            _;
        }else{
            require(whitelist[msg.sender],"2");
            lock=true;
            _;
            lock=false;
        }
    }


    // function uniswapV3FlashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function flashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    // function algebraFlashCallback(uint , uint , bytes calldata data) external payable {unchecked{address(this).call(data);} }

    function uniswapV3SwapCallback(int am0 , int am1, bytes calldata) external payable{
        unchecked{
            require(lock,"3");
            IERC20(am0>am1?IUniV3Pool(msg.sender).token0():IUniV3Pool(msg.sender).token1()).transfer(msg.sender,uint(am0>am1?am0:am1));
        }
    }

    // function swapCallback(int, int, bytes calldata data) external payable {unchecked{address(this).call(data);} }

    function algebraSwapCallback(int am0, int am1, bytes calldata) external payable {
        unchecked{
            require(lock,"3");
            IERC20(am0>am1?IUniV3Pool(msg.sender).token0():IUniV3Pool(msg.sender).token1()).transfer(msg.sender,uint(am0>am1?am0:am1));
        }
    }

    // function apeCall(address, uint, uint, bytes calldata data) external payable {unchecked{address(this).call(data);} }

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